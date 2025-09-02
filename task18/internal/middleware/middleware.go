package middleware

import (
	"calendar/logger"
	"context"
	"github.com/google/uuid"
	"net/http"
	"time"
)

const RequestIDKey ctxKey = "requestID"

type ctxKey string

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(RequestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func CharsetMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := GetRequestID(r.Context())

		lw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lw, r)

		duration := time.Since(start)

		log := logger.RequestLogger.With(
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status_code", lw.statusCode,
			"duration_ms", duration.String(),
			"content_length", r.ContentLength,
		)

		log.Info("HTTP request")
	})
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}
