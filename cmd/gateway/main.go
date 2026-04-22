package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"auto-http-fetcher/internal/core/config"
	pb "auto-http-fetcher/proto/fetcher/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @Title						Auto HTTP Fetcher API
// @Description					Единая точка входа в сервис: авторизация, модули, вебхуки, аналитика и ручной запуск HTTP-запросов.
// @Version						1.0
// @Servers.Url					/api/v1
// @SecurityDefinitions.APIKey	Bearer
// @In							header
// @Name						Authorization
// @Description					Формат: `Bearer {token}`
func main() {
	mainMux := http.NewServeMux()
	RegisterSwagger(mainMux)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	analyticsAddr := config.Get("ANALYTICS_ADDR", "localhost:8080")
	mainMux.HandleFunc("/api/v1/analytics", func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   analyticsAddr,
		})
		proxy.ServeHTTP(w, r)
	})

	modulesAddr := config.Get("MODULES_ADDR", "localhost:8090")
	mainMux.HandleFunc("/api/v1/module/", func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   modulesAddr,
		})
		proxy.ServeHTTP(w, r)
	})

	mainMux.HandleFunc("/api/v1/modules/", func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   modulesAddr,
		})
		proxy.ServeHTTP(w, r)
	})

	usersAddr := config.Get("USERS_ADDR", "localhost:8095")
	mainMux.HandleFunc("/api/v1/auth/", func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   usersAddr,
		})
		proxy.ServeHTTP(w, r)
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
