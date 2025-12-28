package cmd

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"aegis/pkg/ai"
	"aegis/pkg/config"
	"aegis/pkg/server"
	"aegis/pkg/server/handlers"
)

func RunWebServer(opts config.Options, port int, assets embed.FS) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root (current euid=%d)", os.Geteuid())
	}

	app := server.NewApp(opts)
	app.SetReady()

	go func() {
		if err := app.Run(); err != nil {
			log.Printf("Tracer error: %v", err)
		}
	}()

	mux := http.NewServeMux()
	registerAPI(mux, app)
	registerStatic(mux, assets)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		ai.StopOllamaRuntime()
		server.Shutdown(context.Background())
	}()

	log.Printf("========================================")
	log.Printf("Aegis Web UI: http://localhost:%d", port)
	log.Printf("========================================")

	return server.ListenAndServe()
}

func registerStatic(mux *http.ServeMux, assets embed.FS) {
	frontendFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		log.Printf("Warning: frontend assets not found: %v", err)
		return
	}
	mux.Handle("/", http.FileServer(http.FS(frontendFS)))
}

func registerAPI(mux *http.ServeMux, app *server.App) {
	// Register all API handlers
	handlers.RegisterStatsHandlers(mux, app)
	handlers.RegisterEventsHandlers(mux, app)
	handlers.RegisterRulesHandlers(mux, app)
	handlers.RegisterAIHandlers(mux, app)
	handlers.RegisterSettingsHandlers(mux, app)
	handlers.RegisterQueryHandlers(mux, app)
}
