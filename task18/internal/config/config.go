package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Level        string
	Port         string
	LogFilePath  string
	LogToFile    string
	ReadTimeOut  time.Duration
	WriteTimeOut time.Duration
	IdleTimeOut  time.Duration
}

func LoadCfg() *Config {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file", err)
		panic(err)
	}

	cfg := &Config{
		Level:        os.Getenv("LEVEL"),
		Port:         os.Getenv("PORT"),
		LogFilePath:  os.Getenv("LOG_FILE_PATH"),
		LogToFile:    os.Getenv("LOG_TO_FILE"),
		ReadTimeOut:  parseDuration(os.Getenv("READ_TIMEOUT")),
		WriteTimeOut: parseDuration(os.Getenv("WRITE_TIMEOUT")),
		IdleTimeOut:  parseDuration(os.Getenv("IDLE_TIMEOUT")),
	}

	return cfg
}

func parseDuration(durationStr string) time.Duration {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		times, _ := strconv.Atoi(durationStr)
		return time.Duration(times) * time.Second
	}

	return duration
}
