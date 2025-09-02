package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

const (
	URLArg = 0
)

type Config struct {
	URL          string
	OutputDir    string
	UserAgent    string
	CountDepth   int
	CountWorkers int
	TimeOut      time.Duration
	Delay        time.Duration
	SameDomain   bool
	RobotsTxt    bool
}

func ParseFlag() *Config {
	cfg := &Config{
		OutputDir:    "./download",
		UserAgent:    "Mozilla/5.0",
		CountDepth:   1,
		CountWorkers: 5,
		TimeOut:      30 * time.Second,
		Delay:        100 * time.Millisecond,
		SameDomain:   true,
		RobotsTxt:    true,
	}

	flag.StringVar(&cfg.OutputDir, "o", cfg.OutputDir, "output directory")
	flag.StringVar(&cfg.UserAgent, "ua", cfg.UserAgent, "User-Agent string")
	flag.IntVar(&cfg.CountDepth, "d", cfg.CountDepth, "recursion depth")
	flag.IntVar(&cfg.CountWorkers, "w", cfg.CountWorkers, "number concurrent workers")
	flag.DurationVar(&cfg.TimeOut, "t", cfg.TimeOut, "request timeout")
	flag.DurationVar(&cfg.Delay, "delay", cfg.Delay, "delay between requests")
	flag.BoolVar(&cfg.SameDomain, "s", cfg.SameDomain, "only download from same domain")
	flag.BoolVar(&cfg.RobotsTxt, "r", cfg.RobotsTxt, "detect robots.txt")

	flag.Usage = func() {
		fmt.Print("Error with args \nExample use: go run main.go https://example.com\n")
	}

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	cfg.URL = flag.Arg(URLArg)
	return cfg
}
