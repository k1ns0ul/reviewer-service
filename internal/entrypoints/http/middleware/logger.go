package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"reviewer-service/pkg/logger"

	"github.com/google/uuid"
)

func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Генерируем request_id
			requestID := uuid.New().String()

			// Добавляем logger с request_id в context
			logWithID := logger.WithRequestID(log, requestID)
			ctx := logger.WithLogger(r.Context(), logWithID)
			r = r.WithContext(ctx)

			// Логируем входящий запрос
			logWithID.Info("incoming request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)

			// Оборачиваем ResponseWriter для захвата статус кода
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Выполняем запрос
			next.ServeHTTP(wrapped, r)

			// Логируем результат
			duration := time.Since(start)
			logWithID.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
