package main

import (
	"net/http"

	autofetcher "auto-http-fetcher"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func RegisterSwagger(mux *http.ServeMux) {
	mux.HandleFunc("/swagger/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, autofetcher.OpenAPIFS, "api/openapi.json")
	})
	mux.HandleFunc("/swagger/openapi.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, autofetcher.OpenAPIFS, "api/openapi.yml")
	})
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/openapi.json"),
	))
}
