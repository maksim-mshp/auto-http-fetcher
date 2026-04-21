package main

import (
	"auto-http-fetcher/internal/core/di"

	"context"
	"fmt"
	"log"
)

func main() {
	ctx := context.Background()

	app, err := di.NewModulesApp(ctx)
	if err != nil {
		log.Println(fmt.Sprintf("error initializing module service: %v", err))
		return
	}
	if err = app.Start(ctx); err != nil {
		log.Printf("error starting app: %v", err)
	}
	log.Println("app shutdown gracefully")
}
