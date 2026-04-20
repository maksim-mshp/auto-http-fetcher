package main

import (
	"auto-http-fetcher/internal/core/config"
	"auto-http-fetcher/internal/core/logger"
	grpcInfra "auto-http-fetcher/internal/response/infra/grpc"
	pgInfra "auto-http-fetcher/internal/response/infra/postgres"
	"auto-http-fetcher/internal/response/service"
	pb "auto-http-fetcher/proto/fetcher/v1"
	"context"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	if err := config.LoadDotEnv(".env"); err != nil {
		panic(err)
	}
	postgresURL := config.MustGet("POSTGRES_URL")
	grpcPort := config.Get("GRPC_PORT", ":50051")
	env := config.Get("ENV", "Development")

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		panic(err)
	}

	repo := pgInfra.NewResponseRepo(pool)
	logger := logger.New(env)

	fetcher := service.NewFetcher(repo, logger)
	grpcServer := grpc.NewServer()
	handler := grpcInfra.NewHandler(fetcher)

	pb.RegisterFetcherServiceServer(grpcServer, handler)

	lis, err := net.Listen("tcp", grpcPort)

	if err != nil {
		panic(err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}

}
