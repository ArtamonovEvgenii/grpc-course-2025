package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	app "github.com/ArtamonovEvgenii/grpc-course-2025/internal/app/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	err := app.Run(ctx)
	if err != nil {
		os.Exit(1)
	}
}
