package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"aegis/pkg/ai"
	"aegis/pkg/ai/types"
	"aegis/pkg/events"
	"aegis/pkg/proc"
	"aegis/pkg/rules"
	"aegis/pkg/server"
	"aegis/pkg/storage"
)

func RegisterAIHandlers(mux *http.ServeMux, app *server.App) {
	mux.HandleFunc("/api/ai/status", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		var status types.StatusDTO
		if app.AIService() != nil {
			status = app.AIService().GetStatus()
		} else {
			status = types.StatusDTO{Status: "unavailable"}
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
			json.NewEncoder(w).Encode([]types.Message{})
			return
		}

		history := app.GetChatHistory(sessionID)
		if history == nil {
			history = []types.Message{}
		}

		json.NewEncoder(w).Encode(history)
	})

	mux.HandleFunc("/api/ai/generate-rule", func(w http.ResponseWriter, r *http.Request) {
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

		var req types.RuleGenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
		defer cancel()

		core := app.Core()
		if core == nil {
			http.Error(w, "Core components not available", http.StatusServiceUnavailable)
			return
		}

		response, err := app.AIService().GenerateRule(ctx, &req, core.RuleEngine, core.Storage)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/api/ai/explain", func(w http.ResponseWriter, r *http.Request) {
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

		// Accept both camelCase and snake_case payloads
		var raw map[string]any
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		var req types.ExplainRequest
		if v, ok := raw["eventId"].(string); ok {
			req.EventID = v
		} else if v, ok := raw["event_id"].(string); ok {
			req.EventID = v
		}
		if v, ok := raw["question"].(string); ok {
			req.Question = v
		}
		if v, ok := raw["eventData"].(map[string]any); ok {
			req.EventData = &storage.Event{Data: v, Timestamp: time.Now()}
		} else if v, ok := raw["event_data"].(map[string]any); ok {
			req.EventData = &storage.Event{Data: v, Timestamp: time.Now()}
		}

		// Build storage.Event from frontend shape
		var event *storage.Event
		if req.EventData != nil {
			event = &storage.Event{Timestamp: req.EventData.Timestamp, Data: req.EventData.Data}
			// Try to infer type from eventData.type
			if m, ok := req.EventData.Data.(map[string]any); ok {
				// Timestamp may be at top-level or under header.timestamp (ms)
				if ts, ok := m["timestamp"].(float64); ok {
					event.Timestamp = time.UnixMilli(int64(ts))
				} else if hdr, ok := m["header"].(map[string]any); ok {
					if ts, ok := hdr["timestamp"].(float64); ok {
						event.Timestamp = time.UnixMilli(int64(ts))
					}
				}
				// Type is normalized as lowercase string
				if t, ok := m["type"].(string); ok {
					switch t {
					case "exec":
						event.Type = events.EventTypeExec
					case "file":
						event.Type = events.EventTypeFileOpen
					case "connect":
						event.Type = events.EventTypeConnect
					}
				}
			}
		}

		if event == nil {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		core := app.Core()
		if core == nil {
			http.Error(w, "Core components not available", http.StatusServiceUnavailable)
			return
		}

		var processTree *proc.ProcessTree
		if core.ProcessTree != nil {
			processTree = core.ProcessTree
		}
		response, err := app.AIService().ExplainEvent(ctx, &req, event, core.RuleEngine, core.Storage, core.ProfileReg, core.WorkloadReg, processTree, app.Stats())
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/api/ai/analyze", func(w http.ResponseWriter, r *http.Request) {
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

		var req types.AnalyzeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		core := app.Core()
		if core == nil {
			http.Error(w, "Core components not available", http.StatusServiceUnavailable)
			return
		}

		var processTree *proc.ProcessTree
		if core.ProcessTree != nil {
			processTree = core.ProcessTree
		}
		var store storage.EventStore
		if core.Storage != nil {
			store = core.Storage
		}
		response, err := app.AIService().Analyze(ctx, &req, core.ProfileReg, core.WorkloadReg, core.RuleEngine, app.Stats(), store, processTree)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/api/ai/sentinel/stream", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		}

		// Subscribe to Sentinel insights
		sentinel := app.Sentinel()
		if sentinel == nil {
			// Sentinel not available, send heartbeat only
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-r.Context().Done():
					return
				case <-ticker.C:
					fmt.Fprintf(w, "data: {\"type\":\"heartbeat\"}\n\n")
					flusher.Flush()
				}
			}
		}

		// Subscribe to insights
		insightCh := sentinel.Subscribe()
		defer func() {
			// Close the channel and let Sentinel handle cleanup
			// Note: We can't directly unsubscribe with receive-only channel
			// Sentinel will handle cleanup when channel is closed
		}()

		// Send existing insights first
		existingInsights := sentinel.GetInsights(10)
		for _, insight := range existingInsights {
			data, _ := json.Marshal(insight)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}

		// Send heartbeat every 30 seconds
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case insight := <-insightCh:
				// New insight received
				data, _ := json.Marshal(insight)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			case <-ticker.C:
				// Send heartbeat
				fmt.Fprintf(w, "data: {\"type\":\"heartbeat\"}\n\n")
				flusher.Flush()
			}
		}
	})

	mux.HandleFunc("/api/ai/sentinel/insights", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		limit := 50
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		// Get insights from Sentinel
		sentinel := app.Sentinel()
		var insights []*ai.Insight
		if sentinel != nil {
			insights = sentinel.GetInsights(limit)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"insights": insights,
			"total":    len(insights),
		})
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

	mux.HandleFunc("/api/ai/review", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

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
			Rule rules.Rule `json:"rule"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if app.AIService() == nil || !app.AIService().IsEnabled() {
			http.Error(w, "AI service not available", http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		// Use Analyze endpoint to review the rule
		analyzeReq := types.AnalyzeRequest{
			Type: "rule",
			ID:   req.Rule.Name,
		}

		core := app.Core()
		if core == nil {
			http.Error(w, "Core components not available", http.StatusServiceUnavailable)
			return
		}

		var processTree *proc.ProcessTree
		if core.ProcessTree != nil {
			processTree = core.ProcessTree
		}
		var store storage.EventStore
		if core.Storage != nil {
			store = core.Storage
		}
		response, err := app.AIService().Analyze(ctx, &analyzeReq, core.ProfileReg, core.WorkloadReg, core.RuleEngine, app.Stats(), store, processTree)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/api/ai/sentinel/action", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

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
			InsightID string         `json:"insight_id"`
			ActionID  string         `json:"action_id"`
			Params    map[string]any `json:"params"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Execute action based on action_id
		switch req.ActionID {
		case "promote":
			// Promote testing rule
			ruleName, ok := req.Params["rule_name"].(string)
			if !ok {
				http.Error(w, "rule_name parameter required for promote action", http.StatusBadRequest)
				return
			}

			if err := app.PromoteRule(ruleName); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"message": fmt.Sprintf("Rule %s promoted successfully", ruleName),
			})

		case "dismiss":
			// Dismiss insight (no-op for now, just acknowledge)
			json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"message": "Insight dismissed",
			})

		case "investigate":
			// Investigate event (could trigger deeper analysis)
			json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"message": "Investigation started",
			})

		default:
			http.Error(w, fmt.Sprintf("Unknown action: %s", req.ActionID), http.StatusBadRequest)
			return
		}
	})
}
