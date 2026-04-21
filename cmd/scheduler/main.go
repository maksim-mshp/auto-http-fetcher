package main

import (
	"auto-http-fetcher/internal/core/di"
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()

	app, err := di.NewSchedulerApp(ctx)
	if err != nil {
		panic(fmt.Sprintf("main.main: %s", err.Error()))
	}
	if err = app.Run(ctx); err != nil {
		panic(fmt.Sprintf("main.main: %s", err.Error()))
	}
}
