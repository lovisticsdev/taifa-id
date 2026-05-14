package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"taifa-id/internal/platform/ids"
)

const CorrelationIDHeader = "X-Correlation-ID"

type contextKey string

const correlationIDContextKey contextKey = "correlation_id"

func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get(CorrelationIDHeader)
		if correlationID == "" {
			correlationID = ids.New("corr")
		}

		w.Header().Set(CorrelationIDHeader, correlationID)

		ctx := context.WithValue(r.Context(), correlationIDContextKey, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CorrelationIDFromContext(ctx context.Context) string {
	value, ok := ctx.Value(correlationIDContextKey).(string)
	if !ok || value == "" {
		return ""
	}

	return value
}

func RequestLoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()

			recorder := &statusRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(recorder, r)

			logger.Info(
				"http request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.statusCode,
				"duration_ms", time.Since(startedAt).Milliseconds(),
				"correlation_id", CorrelationIDFromContext(r.Context()),
			)
		})
	}
}

func RecovererMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error(
						"panic recovered",
						"panic", recovered,
						"stack", string(debug.Stack()),
						"correlation_id", CorrelationIDFromContext(r.Context()),
					)

					WriteError(
						w,
						r,
						http.StatusInternalServerError,
						"INTERNAL_ERROR",
						"An internal error occurred.",
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
