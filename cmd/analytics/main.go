package main

import (
	httpInfra "auto-http-fetcher/internal/analytics/infra/http"
	pgInfra "auto-http-fetcher/internal/analytics/infra/postgres"
	"auto-http-fetcher/internal/analytics/service"
	"auto-http-fetcher/internal/core/config"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := config.LoadDotEnv(".env"); err != nil {
		panic(err)
	}

	postgresURL := config.MustGet("POSTGRES_URL")
	redisURL := config.MustGet("REDIS_URL")
	analyticsPort := config.Get("ANALYTICS_PORT", ":8080")
	redisTTL := config.Get("ANALYTICS_TTL", "60")

	ttl, err := strconv.Atoi(redisTTL)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		panic(err)
	}

	repo := pgInfra.NewAnalyticsRepo(pool)

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	analyticsService := service.NewAnalyticsService(repo, redisClient, time.Duration(ttl)*time.Second)

	handler := httpInfra.NewHandler(analyticsService)
	router := http.NewServeMux()
	handler.RegisterRoutes(router)

	httpServer := http.Server{
		Addr:    analyticsPort,
		Handler: router,
	}

	if err := httpServer.ListenAndServe(); err != nil {
		panic(err)
	}
}
