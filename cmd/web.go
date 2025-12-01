//go:build web

// EulerGuard Web Server
package main

import (
	"context"
	"embed"
	"log"
	"time"

	"eulerguard/pkg/ai"
	"eulerguard/pkg/config"
	"eulerguard/pkg/ui"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	opts := config.ParseOptions()

	prewarmAIRuntime(opts)

	log.Println("Starting EulerGuard Web Server...")
	if err := ui.RunWebServer(opts, opts.WebPort, assets); err != nil {
		log.Fatalf("eulerguard-web: %v", err)
	}
}

func prewarmAIRuntime(opts config.Options) {
	if !opts.AI.Enabled || opts.AI.Mode != "ollama" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	if err := ai.EnsureOllamaRuntime(ctx, opts.AI.Ollama.Model, opts.AI.Ollama.Endpoint); err != nil {
		log.Printf("[AI] Warning: failed to ensure Ollama runtime: %v", err)
	}
}
