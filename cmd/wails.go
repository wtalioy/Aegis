//go:build wails

// EulerGuard Native GUI
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
	ui.WailsAssets = assets
	if err := ui.RunWails(config.ParseOptions()); err != nil {
		log.Fatalf("eulerguard-gui: %v", err)
	}
}
