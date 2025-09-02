package cmd

import (
	"calendar/internal/calendar"
	"calendar/internal/config"
	"calendar/internal/event/repository"
	"calendar/internal/handlers"
	"calendar/internal/server"
	"calendar/logger"

	_ "calendar/docs"
)

// StartService
// @title Calendar Events
// @version 1.0
// @description Сервис для управления событиями календаря
// @host http://localhost:8080
// @BasePath /
// @schemes http
func StartService() {
	cfg := config.LoadCfg()
	logger.InitLogger(cfg.Level, cfg.LogToFile, cfg.LogFilePath)

	eventRepository := repository.NewEventRepository(logger.AppLogger)
	serviceCalendar := calendar.NewServiceCalendar(eventRepository, logger.AppLogger)
	handler := handlers.NewHandlers(serviceCalendar, logger.AppLogger)

	serv := server.NewServer(handler, cfg, logger.AppLogger)

	logger.AppLogger.Info("starting server",
		"on port", cfg.Port,
		"path log file", cfg.LogFilePath,
		"swagger_url", "http://localhost:"+cfg.Port+"/swagger/index.html",
	)

	if err := serv.Start(); err != nil {
		logger.AppLogger.Error("server failed start", "error", err)
	}
}
