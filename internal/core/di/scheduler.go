package di

import (
	"auto-http-fetcher/internal/core/closer"
	"auto-http-fetcher/internal/core/config"
	"auto-http-fetcher/internal/core/logger"
	corePostgres "auto-http-fetcher/internal/core/postgres"
	responsePG "auto-http-fetcher/internal/response/infra/postgres"
	schedulerGrpc "auto-http-fetcher/internal/scheduler/infra/grpc"
	schedulerKafka "auto-http-fetcher/internal/scheduler/infra/kafka"
	schedulerPG "auto-http-fetcher/internal/scheduler/infra/postgres"
	schedulerService "auto-http-fetcher/internal/scheduler/service"
	fetcherpb "auto-http-fetcher/proto/fetcher/v1"
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SchedulerApp struct {
	scheduler *schedulerService.Scheduler
	consumer  *schedulerKafka.Consumer
	logger    *slog.Logger
	closer    *closer.Closer
	pool      *pgxpool.Pool
	grpcConn  *grpc.ClientConn
}

func NewSchedulerApp(ctx context.Context) (*SchedulerApp, error) {
	if err := config.LoadDotEnv(".env"); err != nil {
		return nil, err
	}

	env := config.Get("ENV", "Development")
	postgresURL := config.MustGet("POSTGRES_URL")
	kafkaBroker := config.MustGet("KAFKA_BROKER")
	kafkaTopic := config.MustGet("KAFKA_TOPIC_SCHEDULE_REQUEST")
	kafkaGroup := config.Get("KAFKA_CONSUMER_GROUP", "auto-http-fetcher-scheduler")
	fetcherAddr := config.Get("FETCHER_GRPC_ADDR", "localhost:50051")

	logs := logger.New(env).With("service", "scheduler")
	clsr := closer.New(logs)

	pool, err := corePostgres.Open(ctx, postgresURL)
	if err != nil {
		return nil, err
	}

	grpcConn, err := grpc.NewClient(fetcherAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		pool.Close()
		return nil, err
	}

	fetcher := schedulerGrpc.NewFetcher(fetcherpb.NewFetcherServiceClient(grpcConn))
	scheduler := schedulerService.NewScheduler(logs, fetcher)
	scheduler.SetResponseSaver(responsePG.NewResponseRepo(pool))

	webhookLoader := schedulerPG.NewWebhookLoader(pool, logs)
	webhooks, err := webhookLoader.Load(ctx)
	if err != nil {
		_ = grpcConn.Close()
		pool.Close()
		return nil, err
	}
	for _, webhook := range webhooks {
		scheduler.AddWebhook(webhook)
	}
	logs.Info("loaded scheduled webhooks", "count", len(webhooks))

	consumer, err := schedulerKafka.NewConsumer(
		splitCSV(kafkaBroker),
		kafkaGroup,
		[]string{kafkaTopic},
		scheduler,
		logs,
	)
	if err != nil {
		_ = grpcConn.Close()
		pool.Close()
		return nil, err
	}

	app := &SchedulerApp{
		scheduler: scheduler,
		consumer:  consumer,
		logger:    logs,
		closer:    clsr,
		pool:      pool,
		grpcConn:  grpcConn,
	}

	if err = app.registerClosers(); err != nil {
		_ = consumer.Close()
		_ = grpcConn.Close()
		pool.Close()
		return nil, err
	}

	return app, nil
}

func (a *SchedulerApp) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 1)
	go a.scheduler.Work(runCtx)
	go func() {
		if err := a.consumer.Run(runCtx); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	a.logger.Info("scheduler started")
	select {
	case err := <-errCh:
		cancel()
		return err
	case sig := <-sigCh:
		a.logger.Info("shutdown signal received", "signal", sig)
	}

	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	return a.closer.Close(shutdownCtx)
}

func (a *SchedulerApp) registerClosers() error {
	if err := a.closer.Add("kafka consumer", func(context.Context) error {
		return a.consumer.Close()
	}); err != nil {
		return err
	}

	if err := a.closer.Add("grpc fetcher connection", func(context.Context) error {
		return a.grpcConn.Close()
	}); err != nil {
		return err
	}

	return a.closer.Add("postgres", func(context.Context) error {
		a.pool.Close()
		return nil
	})
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
