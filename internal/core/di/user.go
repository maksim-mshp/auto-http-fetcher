package di

import (
	"auto-http-fetcher/internal/core/closer"
	"auto-http-fetcher/internal/core/config"
	"auto-http-fetcher/internal/core/logger"
	"auto-http-fetcher/internal/core/postgres"
	"auto-http-fetcher/internal/core/security"
	userHandles "auto-http-fetcher/internal/user/infra/http"
	userPG "auto-http-fetcher/internal/user/infra/postgres"
	userService "auto-http-fetcher/internal/user/service"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type UsersApp struct {
	httpServer *http.Server
	logger     *slog.Logger
	closer     *closer.Closer
}

func NewUsersApp(ctx context.Context) (*UsersApp, error) {
	if err := config.LoadDotEnv(".env"); err != nil {
		return nil, err
	}

	env := config.Get("ENV", "Development")
	httpAddr := config.Get("USERS_PORT", ":8095")
	postgresURL := config.MustGet("POSTGRES_URL")

	jwtSecret := config.MustGet("JWT_SECRET")
	jwtTTL, err := time.ParseDuration(config.Get("JWT_TTL", "5h"))
	if err != nil {
		return nil, err
	}

	logs := logger.New(env)
	logs = logs.With("service", "users")
	clsr := closer.New(logs)

	pool, err := postgres.Open(ctx, postgresURL)
	if err != nil {
		return nil, err
	}

	userRepo := userPG.NewUserRepo(pool)

	jwt := security.NewJWTService(jwtSecret, jwtTTL*time.Hour)
	userServ := userService.NewUserService(logs, jwt, userRepo)

	userHandlers := userHandles.NewUserHandlers(logs, userServ)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: userHandles.GetUserRouter(logs, userHandlers),
	}

	err = clsr.Add("postgres", func(_ context.Context) error {
		pool.Close()
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = clsr.Add("users server", func(ctx context.Context) error {
		return server.Shutdown(ctx)
	})
	if err != nil {
		return nil, err
	}

	return &UsersApp{logger: logs, closer: clsr, httpServer: server}, nil
}

func (app *UsersApp) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		app.logger.Info("starting http server", "port", app.httpServer.Addr)
		if err := app.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.logger.Error("http server failed", "error", err)
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		app.logger.Info("shutdown signal received", "signal", sig)
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return nil
}

func (app *UsersApp) Shutdown(ctx context.Context) error {
	return app.closer.Close(ctx)
}
