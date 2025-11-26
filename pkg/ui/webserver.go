package ui

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"eulerguard/pkg/config"
)

func RunWebServer(opts config.Options, port int, assets embed.FS) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root (current euid=%d)", os.Geteuid())
	}

	app := NewApp(opts)
	close(app.ready)

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
		server.Shutdown(context.Background())
	}()

	log.Printf("========================================")
	log.Printf("  EulerGuard Web UI: http://localhost:%d", port)
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

func registerAPI(mux *http.ServeMux, app *App) {
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")
		s := app.GetSystemStats()
		fmt.Fprintf(w, `{"processCount":%d,"containerCount":%d,"eventsPerSec":%.2f,"alertCount":%d,"probeStatus":"%s"}`,
			s.ProcessCount, s.ContainerCount, s.EventsPerSec, s.AlertCount, s.ProbeStatus)
	})

	mux.HandleFunc("/api/alerts", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")
		alerts := app.GetAlerts()
		if len(alerts) == 0 {
			w.Write([]byte("[]"))
			return
		}
		w.Write([]byte("["))
		for i, a := range alerts {
			if i > 0 {
				w.Write([]byte(","))
			}
			fmt.Fprintf(w, `{"id":"%s","timestamp":%d,"severity":"%s","ruleName":"%s","description":"%s","pid":%d,"processName":"%s","inContainer":%t}`,
				a.ID, a.Timestamp, a.Severity, a.RuleName, a.Description, a.PID, a.ProcessName, a.InContainer)
		}
		w.Write([]byte("]"))
	})

	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		}

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				exec, file, net := app.Stats().Rates()
				fmt.Fprintf(w, "data: {\"exec\":%d,\"file\":%d,\"network\":%d}\n\n", exec, file, net)
				flusher.Flush()
			}
		}
	})

	// Get process ancestors for attack chain visualization
	mux.HandleFunc("/api/ancestors/", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		// Extract PID from URL path
		pidStr := strings.TrimPrefix(r.URL.Path, "/api/ancestors/")
		pid, err := strconv.ParseUint(pidStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid PID", http.StatusBadRequest)
			return
		}

		ancestors := app.GetAncestors(uint32(pid))
		if ancestors == nil {
			w.Write([]byte("[]"))
			return
		}

		data, _ := json.Marshal(ancestors)
		w.Write(data)
	})

	// Get all detection rules
	mux.HandleFunc("/api/rules", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		rules := app.GetRules()
		data, _ := json.Marshal(rules)
		w.Write(data)
	})

	// Get probe stats for Kernel X-Ray
	mux.HandleFunc("/api/probes/stats", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		stats := app.GetProbeStats()
		data, _ := json.Marshal(stats)
		w.Write(data)
	})

	// Learning mode status
	mux.HandleFunc("/api/learning/status", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		status := app.GetLearningStatus()
		data, _ := json.Marshal(status)
		w.Write(data)
	})

	// Start learning mode
	mux.HandleFunc("/api/learning/start", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Duration int `json:"duration"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := app.StartLearning(req.Duration); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	// Stop learning mode
	mux.HandleFunc("/api/learning/stop", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rules, err := app.StopLearning()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(rules)
		w.Write(data)
	})

	// Apply whitelist rules
	mux.HandleFunc("/api/learning/apply", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Indices []int `json:"indices"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := app.ApplyWhitelistRules(req.Indices); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	// SSE stream for all events (LiveStream page)
	mux.HandleFunc("/api/stream", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		}

		// Subscribe to events from the bridge
		events := make(chan any, 100)
		app.Stats().SubscribeEvents(events)
		defer app.Stats().UnsubscribeEvents(events)

		for {
			select {
			case <-r.Context().Done():
				return
			case event := <-events:
				data, _ := json.Marshal(event)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	})
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}
