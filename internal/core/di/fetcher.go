package di

import (
	"auto-http-fetcher/internal/core/closer"
	"auto-http-fetcher/internal/core/config"
	"auto-http-fetcher/internal/core/logger"
	grpcInfra "auto-http-fetcher/internal/response/infra/grpc"
	pgInfra "auto-http-fetcher/internal/response/infra/postgres"
	"auto-http-fetcher/internal/response/service"
	pb "auto-http-fetcher/proto/fetcher/v1"
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type FetcherApp struct {
	grpcServer *grpc.Server
	logger     *slog.Logger
	closer     *closer.Closer
	pool       *pgxpool.Pool
}

func NewFetcherApp(ctx context.Context) (*FetcherApp, error) {
	if err := config.LoadDotEnv(".env"); err != nil {
		return nil, err
	}
	postgresURL := config.MustGet("POSTGRES_URL")
	env := config.Get("ENV", "Development")

	pool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	repo := pgInfra.NewResponseRepo(pool)
	logger := logger.New(env)
	closer := closer.New(logger)
	fetcher := service.NewFetcher(repo, logger)
	grpcServer := grpc.NewServer()
	handler := grpcInfra.NewHandler(fetcher)

	pb.RegisterFetcherServiceServer(grpcServer, handler)

	return &FetcherApp{
		grpcServer: grpcServer,
		logger:     logger,
		closer:     closer,
		pool:       pool,
	}, nil
}

func (f *FetcherApp) Run() error {
	grpcPort := config.Get("GRPC_PORT", ":50051")
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return err
	}

	err = f.closer.Add("listener", func(ctx context.Context) error {
		if err := lis.Close(); err != nil {
			f.logger.Error("failed to close listener", "error", err)
			return err
		}
		f.logger.Info("listener closed")
		return nil
	})

	if err != nil {
		return err
	}

	err = f.closer.Add("grpc_server", func(ctx context.Context) error {
		f.grpcServer.GracefulStop()
		f.logger.Info("gRPC server stopped")
		return nil
	})

	if err != nil {
		return err
	}

	err = f.closer.Add("database", func(ctx context.Context) error {
		f.pool.Close()
		f.logger.Info("Database connection closed")
		return nil
	})

	if err != nil {
		return err
	}

	errCh := make(chan error, 1)

	go func() {
		f.logger.Info("Starting gRPC server", "port", grpcPort)
		if err := f.grpcServer.Serve(lis); err != nil {
			errCh <- err
		}
	}()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		f.logger.Info("shutdown signal received", "signal", sig)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = f.closer.Close(shutdownCtx); err != nil {
		return err
	}
	return nil
}
