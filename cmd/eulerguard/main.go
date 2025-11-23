package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"eulerguard/pkg/app"
	"eulerguard/pkg/config"
)

func main() {
	opts := config.ParseOptions()
	tracer := app.NewExecveTracer(opts)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := tracer.Run(ctx); err != nil {
		log.Fatalf("eulerguard: %v", err)
	}
}
