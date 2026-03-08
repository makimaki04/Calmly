package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func WithLogging(logger *zap.Logger) func(next http.Handler) http.Handler {
	if logger == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	logger = logger.With(zap.String("component", "logger_middleware"))

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				logger.Info("Served",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("proto", r.Proto),
					zap.String("reqId", middleware.GetReqID(r.Context())),
					zap.Duration("duration", time.Since(start)),
					zap.Int("status", ww.Status()),
					zap.Int("size", ww.BytesWritten()),
				)
			}()

			logger.Info("Started request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("proto", r.Proto),
				zap.String("reqId", middleware.GetReqID(r.Context())),
				zap.String("remote", r.RemoteAddr),
			)

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
