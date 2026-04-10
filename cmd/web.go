//go:build web

// Aegis Web Server
package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"aegis/internal/app"
	"aegis/internal/platform/ai/runtime"
	internalconfig "aegis/internal/platform/config"
	httpapi "aegis/internal/platform/http"
)

//go:embed all:assets/frontend/dist
var embeddedAssets embed.FS

func main() {
	cfg, configPath, err := internalconfig.Load(os.Args[1:])
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	prewarmAIRuntime(cfg)

	log.Println("Starting Aegis Web Server...")
	runtimeApp := app.NewRuntime(cfg, configPath)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := runtimeApp.Start(ctx); err != nil {
			log.Printf("Runtime error: %v", err)
		}
	}()
	if err := runWebServer(cfg, runtimeApp); err != nil {
		log.Fatalf("aegis-web: %v", err)
	}
}

func prewarmAIRuntime(cfg internalconfig.Config) {
	if cfg.Analysis.Mode != "ollama" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	if err := runtime.EnsureOllamaRuntime(ctx, cfg.Analysis.Ollama.Model, cfg.Analysis.Ollama.Endpoint); err != nil {
		log.Printf("[AI] Warning: failed to ensure Ollama runtime: %v", err)
	}
}

func runWebServer(cfg internalconfig.Config, runtimeApp *app.Runtime) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root (current euid=%d)", os.Geteuid())
	}

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Server.Port),
		Handler: httpapi.NewHandler(httpapi.DependenciesFromRuntime(runtimeApp), resolveAssets()),
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		runtime.StopOllamaRuntime()
		_ = runtimeApp.Stop(context.Background())
		_ = server.Shutdown(context.Background())
	}()

	log.Printf("========================================")
	log.Printf("Aegis Web UI: http://localhost:%d", cfg.Server.Port)
	log.Printf("========================================")

	return server.ListenAndServe()
}

func resolveAssets() fs.FS {
	if _, err := os.Stat("frontend/dist/index.html"); err == nil {
		return os.DirFS(".")
	}
	assets, err := fs.Sub(embeddedAssets, "assets")
	if err != nil {
		return nil
	}
	return assets
}
