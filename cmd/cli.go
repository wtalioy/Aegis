//go:build !web && !wails

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"eulerguard/pkg/cli"
	"eulerguard/pkg/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cli.RunCLI(config.ParseOptions(), ctx); err != nil {
		log.Fatalf("eulerguard: %v", err)
	}
}
