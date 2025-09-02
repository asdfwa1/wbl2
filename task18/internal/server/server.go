package server

import (
	"calendar/internal/config"
	"calendar/internal/handlers"
	mymiddleware "calendar/internal/middleware"
	"calendar/logger"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	httpServer *http.Server
	handlers   *handlers.Handlers
	config     *config.Config
	log        *slog.Logger
}

func NewServer(handlers *handlers.Handlers, cfg *config.Config, logger *slog.Logger) *Server {
	router := chi.NewRouter()

	router.Use(mymiddleware.RequestIDMiddleware)
	router.Use(mymiddleware.RequestLogger)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(cfg.WriteTimeOut))
	router.Use(mymiddleware.CharsetMiddleware)

	router.Get("/swagger/*", httpSwagger.WrapHandler)

	router.Post("/create_event", handlers.CreateEvent)
	router.Post("/update_event", handlers.UpdateEvent)
	router.Post("/delete_event", handlers.DeleteEvent)
	router.Get("/events_for_day", handlers.EventsForDay)
	router.Get("/events_for_week", handlers.EventsForWeek)
	router.Get("/events_for_month", handlers.EventsForMonth)

	router.Get("/health", handlers.HealthCheck)
	router.NotFound(handlers.NotFound)

	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      router,
			ReadTimeout:  cfg.ReadTimeOut,
			WriteTimeout: cfg.WriteTimeOut,
			IdleTimeout:  cfg.IdleTimeOut,
		},
		handlers: handlers,
		config:   cfg,
		log:      logger,
	}
}

func (s *Server) Start() error {
	notify := make(chan os.Signal, 1)
	signal.Notify(notify, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	<-notify
	s.log.Info("Shutting down server gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Error("Server forced to shutdown", "error", err)
		return err
	}

	logger.Close()

	s.log.Info("Server terminated without incident")
	return nil
}
