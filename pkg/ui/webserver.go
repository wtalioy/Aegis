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

	"eulerguard/pkg/ai"
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
		ai.StopOllamaRuntime()
		server.Shutdown(context.Background())
	}()

	log.Printf("========================================")
	log.Printf("EulerGuard Web UI: http://localhost:%d", port)
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
		fmt.Fprintf(w, `{"processCount":%d,"workloadCount":%d,"eventsPerSec":%.2f,"alertCount":%d,"probeStatus":"%s"}`,
			s.ProcessCount, s.WorkloadCount, s.EventsPerSec, s.AlertCount, s.ProbeStatus)
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
			fmt.Fprintf(w, `{"id":"%s","timestamp":%d,"severity":"%s","ruleName":"%s","description":"%s","pid":%d,"processName":"%s","cgroupId":"%s","action":"%s","blocked":%t}`,
				a.ID, a.Timestamp, a.Severity, a.RuleName, a.Description, a.PID, a.ProcessName, a.CgroupID, a.Action, a.Blocked)
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

	mux.HandleFunc("/api/ancestors/", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

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

	mux.HandleFunc("/api/rules", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		rules := app.GetRules()
		data, _ := json.Marshal(rules)
		w.Write(data)
	})

	mux.HandleFunc("/api/workloads", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		if strings.HasPrefix(r.URL.Path, "/api/workloads/") {
			idStr := strings.TrimPrefix(r.URL.Path, "/api/workloads/")
			workload := app.GetWorkload(idStr)
			if workload == nil {
				http.Error(w, "Workload not found", http.StatusNotFound)
				return
			}
			data, _ := json.Marshal(workload)
			w.Write(data)
			return
		}

		workloads := app.GetWorkloads()
		data, _ := json.Marshal(workloads)
		w.Write(data)
	})

	mux.HandleFunc("/api/workloads/", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		idStr := strings.TrimPrefix(r.URL.Path, "/api/workloads/")
		workload := app.GetWorkload(idStr)
		if workload == nil {
			http.Error(w, "Workload not found", http.StatusNotFound)
			return
		}
		data, _ := json.Marshal(workload)
		w.Write(data)
	})

	mux.HandleFunc("/api/probes/stats", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		stats := app.GetProbeStats()
		data, _ := json.Marshal(stats)
		w.Write(data)
	})

	mux.HandleFunc("/api/learning/status", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		status := app.GetLearningStatus()
		data, _ := json.Marshal(status)
		w.Write(data)
	})

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

	mux.HandleFunc("/api/learning/stop", func(w http.ResponseWriter, r *http.Request) {
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

		rules, err := app.StopLearning()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(rules)
		w.Write(data)
	})

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

		events := make(chan any, 100)
		app.Stats().SubscribeEvents(events)
		defer app.Stats().UnsubscribeEvents(events)

		for {
			select {
			case <-r.Context().Done():
				return
			case event := <-events:
				switch e := event.(type) {
				case sseEvent:
					payload, _ := json.Marshal(e.Data)
					if e.Name != "" {
						fmt.Fprintf(w, "event: %s\n", e.Name)
					}
					fmt.Fprintf(w, "data: %s\n\n", payload)
				default:
					data, _ := json.Marshal(event)
					fmt.Fprintf(w, "data: %s\n\n", data)
				}
				flusher.Flush()
			}
		}
	})

	mux.HandleFunc("/api/ai/status", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		var status ai.StatusDTO
		if app.AIService() != nil {
			status = app.AIService().GetStatus()
		} else {
			status = ai.StatusDTO{Enabled: false, Status: "disabled"}
		}

		json.NewEncoder(w).Encode(status)
	})

	mux.HandleFunc("/api/ai/diagnose", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
		defer cancel()

		result, err := app.Diagnose(ctx)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		if result == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "AI service not available"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	mux.HandleFunc("/api/ai/chat", func(w http.ResponseWriter, r *http.Request) {
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
			Message   string `json:"message"`
			SessionID string `json:"sessionId"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Message == "" {
			http.Error(w, "Message is required", http.StatusBadRequest)
			return
		}

		if req.SessionID == "" {
			req.SessionID = generateSessionID()
		}

		ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
		defer cancel()

		result, err := app.Chat(ctx, req.SessionID, req.Message)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		if result == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "AI service not available"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	mux.HandleFunc("/api/ai/chat/stream", func(w http.ResponseWriter, r *http.Request) {
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
			Message   string `json:"message"`
			SessionID string `json:"sessionId"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Message == "" {
			http.Error(w, "Message is required", http.StatusBadRequest)
			return
		}

		if req.SessionID == "" {
			req.SessionID = generateSessionID()
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
		defer cancel()

		tokenChan, err := app.ChatStream(ctx, req.SessionID, req.Message)
		if err != nil {
			data, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			return
		}

		if tokenChan == nil {
			data, _ := json.Marshal(map[string]string{"error": "AI service not available"})
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			return
		}

		for token := range tokenChan {
			data, err := json.Marshal(token)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	})

	mux.HandleFunc("/api/ai/chat/history", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		sessionID := r.URL.Query().Get("sessionId")
		if sessionID == "" {
			json.NewEncoder(w).Encode([]ai.Message{})
			return
		}

		history := app.GetChatHistory(sessionID)
		if history == nil {
			history = []ai.Message{}
		}

		json.NewEncoder(w).Encode(history)
	})

	mux.HandleFunc("/api/ai/chat/clear", func(w http.ResponseWriter, r *http.Request) {
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
			SessionID string `json:"sessionId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.SessionID != "" {
			app.ClearChat(req.SessionID)
		}

		w.WriteHeader(http.StatusOK)
	})
}

func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}
