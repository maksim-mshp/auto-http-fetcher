package grpc

import (
	"auto-http-fetcher/internal/response/service"
	fetcherpb "auto-http-fetcher/proto/fetcher/v1"
)

type Handler struct {
	fetcherpb.UnimplementedFetcherServiceServer
	fetcher *service.Fetcher
}

func NewHandler(fetcher *service.Fetcher) *Handler {
	return &Handler{
		fetcher: fetcher,
	}
}
