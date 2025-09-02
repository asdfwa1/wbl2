package logger

import (
	"io"
	"log/slog"
	"os"
)

var (
	AppLogger     *slog.Logger
	RequestLogger *slog.Logger
	logFile       *os.File
)

func InitLogger(level string, logToFile string, logFilePath string) {
	appHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: levelForEnv(level),
	})
	AppLogger = slog.New(appHandler)

	var requestOutput io.Writer = os.Stdout
	if logToFile == "true" {
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		logFile = file
		requestOutput = io.MultiWriter(os.Stdout, file)
	}

	requestHandler := slog.NewJSONHandler(requestOutput, &slog.HandlerOptions{
		Level: levelForEnv(level),
	})
	RequestLogger = slog.New(requestHandler)

	slog.SetDefault(AppLogger)
}

func levelForEnv(env string) slog.Level {
	switch env {
	case "dev", "local":
		return slog.LevelDebug
	case "prod":
		return slog.LevelInfo
	default:
		return slog.LevelWarn
	}
}

func Close() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}
