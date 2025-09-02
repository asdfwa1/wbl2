package main

import (
	"fmt"
	"os"
	"task16/config"
	"task16/downloader"
	"time"
)

func main() {
	start := time.Now()

	cfg := config.ParseFlag()
	startWithSetting(cfg)

	dl := downloader.NewDownloader(cfg)

	if err := dl.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nDownload completed in %v\n", time.Since(start))
}

func startWithSetting(cfg *config.Config) {
	fmt.Printf("Starting download of: %s\n", cfg.URL)
	fmt.Printf("Output directory: %s\n", cfg.OutputDir)
	fmt.Printf("User-Agent: %s\n", cfg.UserAgent)
	fmt.Printf("Number depth: %d\n", cfg.CountDepth)
	fmt.Printf("Workers: %d\n", cfg.CountWorkers)
	fmt.Printf("Timeout: %v\n", cfg.TimeOut)
	fmt.Printf("Delay: %v\n", cfg.Delay)
	fmt.Printf("Same domain only: %v\n", cfg.SameDomain)
	fmt.Printf("Respect robots.txt: %v\n", cfg.RobotsTxt)
	fmt.Print("---------------------------------------------------------------------------------------------------------------------------------")
}
