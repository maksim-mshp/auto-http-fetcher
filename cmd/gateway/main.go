package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"auto-http-fetcher/internal/core/config"
	pb "auto-http-fetcher/proto/fetcher/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	mainMux := http.NewServeMux()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	analyticsAddr := config.Get("ANALYTICS_ADDR", "localhost:8080")
	mainMux.HandleFunc("/api/v1/analytics", func(w http.ResponseWriter, r *http.Request) {
		url := "http://" + analyticsAddr + "/api/v1/analytics"
		req, err := http.NewRequest(r.Method, url, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header = r.Header

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	grpcMux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	fetcherAddr := config.Get("FETCHER_ADDR", "localhost:50051")
	err := pb.RegisterFetcherServiceHandlerFromEndpoint(ctx, grpcMux, fetcherAddr, opts)
	if err != nil {
		panic(fmt.Sprintf("main.main: %s", err.Error()))
	}

	mainMux.Handle("/", grpcMux)

	logger.Info("API Gateway server started", "port", ":8081")
	if err = http.ListenAndServe(":8081", mainMux); err != nil {
		panic(fmt.Sprintf("main.main: %s", err.Error()))
	}
}
