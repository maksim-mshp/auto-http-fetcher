package di

import (
	"auto-http-fetcher/internal/core/closer"
	"auto-http-fetcher/internal/core/config"
	kafkaProducer "auto-http-fetcher/internal/core/kafka"
	"auto-http-fetcher/internal/core/logger"
	"auto-http-fetcher/internal/core/postgres"
	"auto-http-fetcher/internal/core/security"
	moduleHandlers "auto-http-fetcher/internal/module/infra/http/handlers"
	"auto-http-fetcher/internal/module/infra/http/router"
	deadLetterQueue "auto-http-fetcher/internal/module/infra/kafka/dlq"
	modulePG "auto-http-fetcher/internal/module/infra/postgres"
	moduleService "auto-http-fetcher/internal/module/service"
	webhookHandlers "auto-http-fetcher/internal/webhook/infra/http/handlers"
	webhookPG "auto-http-fetcher/internal/webhook/infra/postgres"
	webhookService "auto-http-fetcher/internal/webhook/service"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"context"
	"log/slog"
	"net/http"
	"time"
)

type ModulesApp struct {
	httpServer *http.Server
	logger     *slog.Logger
	closer     *closer.Closer
}

func NewModulesApp(ctx context.Context) (*ModulesApp, error) {
	if err := config.LoadDotEnv(".env"); err != nil {
		return nil, err
	}

	env := config.Get("ENV", "Development")
	httpAddr := config.Get("MODULES_PORT", ":8090")
	postgresURL := config.MustGet("POSTGRES_URL")
	kafkaBroker := config.MustGet("KAFKA_BROKER")
	kafkaTopic := config.MustGet("KAFKA_TOPIC_SCHEDULE_REQUEST")
	jwtSecret := config.MustGet("JWT_SECRET")
	jwtTTL, err := time.ParseDuration(config.Get("JWT_TTL", "5h"))
	if err != nil {
		return nil, err
	}

	logs := logger.New(env)
	logs = logs.With("service", "modules")
	clsr := closer.New(logs)

	pool, err := postgres.Open(ctx, postgresURL)
	if err != nil {
		return nil, err
	}

	moduleRepo := modulePG.NewPGModuleRepo(pool)
	webhookRepo := webhookPG.NewPGWebhookRepo(pool)

	kafka, err := kafkaProducer.NewProducer([]string{kafkaBroker}, kafkaTopic)
	if err != nil {
		return nil, err
	}

	dlq := deadLetterQueue.NewDeadLetterQueue(logs, kafka)

	moduleServ := moduleService.NewModuleService(logs, kafka, dlq, moduleRepo)
	webhookServ := webhookService.NewWebhookService(logs, kafka, dlq, webhookRepo)
	jwt := security.NewJWTService(jwtSecret, jwtTTL*time.Hour)

	moduleHandles := moduleHandlers.NewModuleHandlers(logs, *moduleServ)
	webhookHandles := webhookHandlers.NewWebhookHandlers(logs, *webhookServ)

	moduleRouter := router.GetModulesRouter(logs, jwt, moduleHandles, webhookHandles)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: moduleRouter,
	}

	err = clsr.Add("postgres", func(_ context.Context) error {
		pool.Close()
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = clsr.Add("kafka producer", func(_ context.Context) error {
		return kafka.Close()
	})
	if err != nil {
		return nil, err
	}
	err = clsr.Add("module server", func(ctx context.Context) error {
		return server.Shutdown(ctx)
	})
	if err != nil {
		return nil, err
	}

	return &ModulesApp{logger: logs, closer: clsr, httpServer: server}, nil
}

func (app *ModulesApp) Start(ctx context.Context) error {

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

func (app *ModulesApp) Shutdown(ctx context.Context) error {
	return app.closer.Close(ctx)
}
