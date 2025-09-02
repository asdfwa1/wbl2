package parser

import (
	"net/url"
	"regexp"
	"strings"
)

type ParserHTML struct {
	lintPatterns     map[string]*regexp.Regexp
	resourcePatterns map[string]*regexp.Regexp
}

func NewParserHTML() *ParserHTML {
	return &ParserHTML{
		lintPatterns: map[string]*regexp.Regexp{
			"a":    regexp.MustCompile(`<a[^>]+href=["']([^"']+)["']`),
			"area": regexp.MustCompile(`<area[^>]+href=["']([^"']+)["']`),
		},
		resourcePatterns: map[string]*regexp.Regexp{
			"link":   regexp.MustCompile(`<link[^>]+href=["']([^"']+)["']`),
			"script": regexp.MustCompile(`<script[^>]+src=["']([^"']+)["']`),
			"img":    regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`),
			"srcset": regexp.MustCompile(`srcset=["']([^"']+)["']`),
			"source": regexp.MustCompile(`<source[^>]+src=["']([^"']+)["']`),
			"video":  regexp.MustCompile(`<video[^>]+src=["']([^"']+)["']`),
			"audio":  regexp.MustCompile(`<audio[^>]+src=["']([^"']+)["']`),
			"iframe": regexp.MustCompile(`<iframe[^>]+src=["']([^"']+)["']`),
			"embed":  regexp.MustCompile(`<embed[^>]+src=["']([^"']+)["']`),
			"object": regexp.MustCompile(`<object[^>]+data=["']([^"']+)["']`),
			"icon":   regexp.MustCompile(`<link[^>]+rel=["']icon["'][^>]+href=["']([^"']+)["']`),
		},
	}
}

func (p *ParserHTML) ExtractLinksAndResources(baseUrl *url.URL, content []byte) ([]string, []string, error) {
	links := p.extractPatterns(baseUrl, content, p.lintPatterns)
	resources := p.extractPatterns(baseUrl, content, p.resourcePatterns)

	cssUrls := p.extractCssUrls(baseUrl, content)
	resources = append(resources, cssUrls...)

	return p.unique(links), p.unique(resources), nil
}

func (p *ParserHTML) extractPatterns(baseUrl *url.URL, content []byte, patterns map[string]*regexp.Regexp) []string {
	var res []string

	for _, pattern := range patterns {
		matches := pattern.FindAllSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				link := string(match[1])
				if pattern == p.resourcePatterns["srcset"] {
					urls := p.parseSrcset(baseUrl, link)
					res = append(res, urls...)
				} else {
					absoluteURL, err := p.normalizeUrl(baseUrl, link)
					if err == nil && absoluteURL != "" {
						res = append(res, absoluteURL)
					}
				}
			}
		}
	}

	return res
}

func (p *ParserHTML) extractCssUrls(baseURL *url.URL, content []byte) []string {
	var urls []string

	cssUrlPattern := regexp.MustCompile(`url\(["']?([^"')]+)["']?\)`)
	matches := cssUrlPattern.FindAllSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			cssUrl := string(match[1])
			if strings.HasPrefix(cssUrl, "data:") {
				continue
			}
			absoluteURL, err := p.normalizeUrl(baseURL, cssUrl)
			if err == nil && absoluteURL != "" {
				urls = append(urls, absoluteURL)
			}
		}
	}

	return urls
}

func (p *ParserHTML) parseSrcset(baseUrl *url.URL, srcset string) []string {
	var urls []string

	parts := strings.Split(srcset, ",")
	for _, part := range parts {
		urlPart := strings.TrimSpace(strings.Split(part, " ")[0])
		if absoluteURL, err := p.normalizeUrl(baseUrl, urlPart); err == nil && absoluteURL != "" {
			urls = append(urls, absoluteURL)
		}
	}
	return urls
}

func (p *ParserHTML) normalizeUrl(baseUrl *url.URL, link string) (string, error) {
	if link == "" || strings.HasPrefix(link, "#") {
		return "", nil
	}

	parsedLink, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	if parsedLink.Fragment != "" {
		parsedLink.Fragment = ""
	}

	if parsedLink.RawQuery != "" && !strings.Contains(link, ".html") {
		parsedLink.RawQuery = ""
	}

	return baseUrl.ResolveReference(parsedLink).String(), nil
}

func (p *ParserHTML) unique(urls []string) []string {
	seen := make(map[string]bool)
	var res []string

	for _, url := range urls {
		if url != "" && !seen[url] {
			seen[url] = true
			res = append(res, url)
		}
	}

	return res
}
