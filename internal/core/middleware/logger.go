package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logger(logs *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logs.Info("Pending request", "client", r.RemoteAddr, "endpoint", r.URL.String(), "method", r.Method,
				"time", time.Now().Format("02.01.2006 15:04:05"))
			next.ServeHTTP(w, r)
		})
	}
}
