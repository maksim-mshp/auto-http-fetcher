package main

import (
	"auto-http-fetcher/internal/core/di"
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	app, err := di.NewAnalyticsApp(ctx)
	if err != nil {
		panic(fmt.Sprintf("main.main: %s", err.Error()))
	}
	err = app.Run()
	if err != nil {
		panic(fmt.Sprintf("main.main: %s", err.Error()))
	}
}
