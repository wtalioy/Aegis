//go:build web

// Aegis Web Server
package main

import (
	"context"
	"embed"
	"log"
	"time"

	"aegis/pkg/ai"
	"aegis/pkg/config"
	"aegis/pkg/server/cmd"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	opts := config.ParseOptions()

	prewarmAIRuntime(opts)

	log.Println("Starting Aegis Web Server...")
	if err := cmd.RunWebServer(opts, opts.WebPort, assets); err != nil {
		log.Fatalf("aegis-web: %v", err)
	}
}

func prewarmAIRuntime(opts config.Options) {
	if opts.AI.Mode != "ollama" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	if err := ai.EnsureOllamaRuntime(ctx, opts.AI.Ollama.Model, opts.AI.Ollama.Endpoint); err != nil {
		log.Printf("[AI] Warning: failed to ensure Ollama runtime: %v", err)
	}
}
