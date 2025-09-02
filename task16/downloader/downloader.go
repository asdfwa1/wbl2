package downloader

import (
	"context"
	"fmt"
	"github.com/temoto/robotstxt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"task16/config"
	"task16/parser"
	"time"
)

type Downloader struct {
	config       *config.Config
	client       *http.Client
	parser       *parser.ParserHTML
	visited      map[string]bool
	visitedMutex sync.Mutex
	wg           sync.WaitGroup
	done         chan struct{}
	queue        chan *downloadTask
}

type downloadTask struct {
	url   string
	depth int
}

func NewDownloader(cfg *config.Config) *Downloader {
	return &Downloader{
		config: cfg,
		client: &http.Client{
			Timeout: cfg.TimeOut,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) > 11 {
					return fmt.Errorf("to many redirects")
				}
				return nil
			},
		},
		parser:  parser.NewParserHTML(),
		visited: make(map[string]bool),
		done:    make(chan struct{}),
		queue:   make(chan *downloadTask, 1000),
	}
}

func (d *Downloader) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	if err := os.MkdirAll(d.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	if d.config.RobotsTxt {
		if err := d.checkRobotsTxt(); err != nil {
			return err
		}
	}

	for i := 0; i < d.config.CountWorkers; i++ {
		d.wg.Add(1)
		go d.workers(ctx)
	}

	d.queue <- &downloadTask{url: d.config.URL, depth: 0}

	go func() {
		d.wg.Wait()
		close(d.done)
	}()

	select {
	case <-ctx.Done():
		fmt.Println("download timeout after 1 minute")
		return nil
	case <-d.done:
		return nil
	}
}

func (d *Downloader) checkRobotsTxt() error {
	baseUrl, err := url.Parse(d.config.URL)
	if err != nil {
		return err
	}

	robotsUrl := fmt.Sprintf("%s://%s/robots.txt", baseUrl.Scheme, baseUrl.Host)

	req, err := http.NewRequest("GET", robotsUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", d.config.UserAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			fmt.Printf("robots.txt not found at %s. Proceeding without rules.\n", robotsUrl)
			return nil
		}
		return fmt.Errorf("failed to fetch robots.txt: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error close body: %v\n", err)
		}
	}()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("\nFound robots.txt at %s\n", robotsUrl)
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read robots.txt content: %v", err)
		}
		robotsData, err := robotstxt.FromBytes(content)
		if err != nil {
			return fmt.Errorf("failed to parse robots.txt: %v", err)
		}
		fmt.Println("Successfully parsed robots.txt")

		if !robotsData.TestAgent(d.config.URL, d.config.UserAgent) {
			fmt.Printf("WARNING: User-agent '%s' is disallowed from accessing %s\n",
				d.config.UserAgent, d.config.URL)
		}

		if err := d.saveRobotsTxt(robotsUrl, content); err != nil {
			fmt.Printf("Error saving robots.txt locally: %v\n", err)
		}
	} else if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("robots.txt not found at %s. Proceeding without rules.\n", robotsUrl)
	} else {
		return fmt.Errorf("unexpected status code for robots.txt: %d", resp.StatusCode)
	}

	return nil
}

func (d *Downloader) saveRobotsTxt(robotsUrl string, content []byte) error {
	if d.config.OutputDir == "" {
		return nil
	}

	if len(content) == 0 {
		fmt.Printf("robots.txt from %s is empty. Skipping save.\n", robotsUrl)
		return nil
	}

	parsedURL, err := url.Parse(robotsUrl)
	if err != nil {
		return err
	}
	filePath := filepath.Join(d.config.OutputDir, parsedURL.Host, "robots.txt")

	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dirPath, err)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write robots.txt to file %s: %v", filePath, err)
	}

	fmt.Printf("Saved robots.txt to %s\n", filePath)
	return nil
}

func (d *Downloader) workers(ctx context.Context) {
	defer d.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-d.queue:
			if !ok {
				return
			}

			if task.depth > d.config.CountDepth {
				continue
			}

			if d.isVisited(task.url) {
				continue
			}
			d.markVisited(task.url)

			time.Sleep(d.config.Delay)

			content, contentType, err := d.downloadResource(task.url)
			if err != nil {
				fmt.Printf("Error downloading %s: %v\n", task.url, err)
				continue
			}

			localPath, err := d.saveToDirectory(task.url, content, contentType)
			if err != nil {
				fmt.Printf("Error saving %s: %v\n", task.url, err)
				continue
			}

			fmt.Printf("Downloaded: %s -> %s\n", task.url, localPath)

			if strings.Contains(contentType, "text/html") {
				baseUrl, _ := url.Parse(task.url)
				links, resources, err := d.parser.ExtractLinksAndResources(baseUrl, content)
				if err != nil {
					fmt.Printf("Error extracting links from %s: %v\n", task.url, err)
					continue
				}

				for _, resource := range resources {
					if !d.isVisited(resource) && d.shouldDownload(resource, baseUrl) {
						select {
						case d.queue <- &downloadTask{url: resource, depth: task.depth}:
						case <-ctx.Done():
							return
						}
					}
				}

				for _, link := range links {
					if task.depth+1 <= d.config.CountDepth && d.shouldDownload(link, baseUrl) {
						select {
						case d.queue <- &downloadTask{url: link, depth: task.depth + 1}:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		case <-d.done:
			close(d.queue)
			return
		}
	}
}

func (d *Downloader) downloadResource(url string) ([]byte, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("User-Agent", d.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru-RU,en-US,en;q=0.5")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error close body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return content, resp.Header.Get("Content-Type"), nil
}

func (d *Downloader) saveToDirectory(urlStr string, content []byte, contentType string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	filePath := path.Join(d.config.OutputDir, parsedURL.Host, parsedURL.Path)

	needsIndexHTML := false

	if strings.HasSuffix(parsedURL.Path, "/") {
		needsIndexHTML = true
	} else if filepath.Ext(parsedURL.Path) == "" {
		if strings.Contains(contentType, "text/html") {
			needsIndexHTML = true
		}
	}

	if needsIndexHTML {
		filePath = path.Join(filePath, "index.html")
	}

	if strings.Contains(filePath, "?") {
		filePath = strings.Split(filePath, "?")[0]
	}

	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %v", dirPath, err)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

	return filePath, nil
}

func (d *Downloader) shouldDownload(urlStr string, baseURL *url.URL) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	if urlStr == "" || strings.HasPrefix(urlStr, "#") {
		return false
	}

	if d.config.SameDomain {
		if parsedURL.Host != "" && parsedURL.Host != baseURL.Host {
			return false
		}
	}

	if parsedURL.Scheme != "" && parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	if strings.Contains(urlStr, "?") && !strings.Contains(urlStr, ".html") {
		return false
	}

	return true
}

func (d *Downloader) isVisited(url string) bool {
	d.visitedMutex.Lock()
	defer d.visitedMutex.Unlock()
	return d.visited[url]
}

func (d *Downloader) markVisited(url string) {
	d.visitedMutex.Lock()
	defer d.visitedMutex.Unlock()
	d.visited[url] = true
}
