package di

import (
	httpInfra "auto-http-fetcher/internal/analytics/infra/http"
	pgInfra "auto-http-fetcher/internal/analytics/infra/postgres"
	"auto-http-fetcher/internal/analytics/service"
	"auto-http-fetcher/internal/core/closer"
	"auto-http-fetcher/internal/core/config"
	"auto-http-fetcher/internal/core/logger"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AnalyticsApp struct {
	httpServer *http.Server
	logger     *slog.Logger
	closer     *closer.Closer
	redis      *redis.Client
	pool       *pgxpool.Pool
}

func NewAnalyticsApp(ctx context.Context) (*AnalyticsApp, error) {
	if err := config.LoadDotEnv(".env"); err != nil {
		return nil, err
	}

	env := config.Get("ENV", "Development")
	httpPort := config.Get("ANALYTICS_PORT", ":8080")
	redisURL := config.MustGet("REDIS_URL")
	redisTTL := config.Get("ANALYTICS_TTL", "60")
	postgresURL := config.MustGet("POSTGRES_URL")

	logger := logger.New(env)
	closer := closer.New(logger)
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	pool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	ttl, err := strconv.Atoi(redisTTL)
	if err != nil {
		return nil, err
	}
	repo := pgInfra.NewAnalyticsRepo(pool)
	analyticsService := service.NewAnalyticsService(repo, redisClient, time.Duration(ttl)*time.Second)

	handler := httpInfra.NewHandler(analyticsService)
	router := http.NewServeMux()
	handler.RegisterRoutes(router)
	httpServer := &http.Server{
		Addr:    httpPort,
		Handler: router,
	}

	return &AnalyticsApp{
		httpServer: httpServer,
		logger:     logger,
		closer:     closer,
		redis:      redisClient,
		pool:       pool,
	}, nil
}

func (a *AnalyticsApp) Run() error {
	err := a.closer.Add("http_server", func(ctx context.Context) error {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.logger.Error("http server closing failed")
			return err
		}
		a.logger.Info("http server stopped")
		return nil
	})
	if err != nil {
		return err
	}

	err = a.closer.Add("database", func(ctx context.Context) error {
		a.logger.Info("database stopped")
		a.pool.Close()
		return nil
	})

	if err != nil {
		return err
	}

	err = a.closer.Add("redis", func(ctx context.Context) error {
		if err = a.redis.Close(); err != nil {
			a.logger.Error("redis closing failed")
			return err
		}
		a.logger.Info("redis stopped")
		return nil
	})

	errCh := make(chan error, 1)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		a.logger.Info("starting http server", "port", a.httpServer.Addr)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("http server failed", "error", err)
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		a.logger.Info("shutdown signal received", "signal", sig)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.closer.Close(shutdownCtx); err != nil {
		return err
	}
	return nil
}
