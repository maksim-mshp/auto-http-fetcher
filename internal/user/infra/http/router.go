package http

import (
	"auto-http-fetcher/internal/core/middleware"
	"log/slog"
	"net/http"
)

func GetUserRouter(logger *slog.Logger, userHandles *UserHandlers) http.Handler {
	user := http.NewServeMux()
	user.HandleFunc("POST /api/v1/auth/login", userHandles.Login)
	user.HandleFunc("POST /api/v1/auth/register", userHandles.Register)

	return middleware.Logger(logger)(user)
}
