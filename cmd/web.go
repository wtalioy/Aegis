//go:build web

// EulerGuard Web Server
package main

import (
	"embed"
	"log"

	"eulerguard/pkg/config"
	"eulerguard/pkg/ui"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	opts := config.ParseOptions()
	log.Println("Starting EulerGuard Web Server...")
	if err := ui.RunWebServer(opts, opts.WebPort, assets); err != nil {
		log.Fatalf("eulerguard-web: %v", err)
	}
}
