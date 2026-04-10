package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"aegis/internal/analysis/types"
)

func registerSentinelRoutes(mux *http.ServeMux, deps Dependencies) {
	registerAliases(mux, []string{"/api/v1/sentinel/stream"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		if deps.Analysis == nil || deps.Analysis.Sentinel() == nil {
			writeErrorString(w, http.StatusServiceUnavailable, "Sentinel not available")
			return
		}
		subscription := deps.Analysis.Sentinel().Subscribe()
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, ok := w.(http.Flusher)
		if !ok {
			writeErrorString(w, http.StatusInternalServerError, "streaming not supported")
			return
		}
		defer subscription.Cancel()

		heartbeat := time.NewTicker(30 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case insight, ok := <-subscription.C:
				if !ok {
					return
				}
				data, _ := json.Marshal(insight)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			case <-heartbeat.C:
				fmt.Fprintf(w, "data: {\"type\":\"heartbeat\"}\n\n")
				flusher.Flush()
			}
		}
	})

	registerAliases(mux, []string{"/api/v1/sentinel/insights"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		if deps.Analysis == nil || deps.Analysis.Sentinel() == nil {
			writeJSON(w, http.StatusOK, map[string]any{"insights": []any{}, "total": 0})
			return
		}
		limit := 50
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		insights := deps.Analysis.Sentinel().GetInsights(limit)
		writeJSON(w, http.StatusOK, map[string]any{"insights": insights, "total": len(insights)})
	})

	registerAliases(mux, []string{"/api/v1/sentinel/ask"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		if deps.Analysis == nil {
			writeErrorString(w, http.StatusServiceUnavailable, "AI service not available")
			return
		}
		var req types.AskInsightRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
		defer cancel()
		result, err := deps.Analysis.AskAboutInsight(ctx, &req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	registerAliases(mux, []string{"/api/v1/sentinel/actions"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		var req struct {
			ActionID string         `json:"action_id"`
			Params   map[string]any `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		switch req.ActionID {
		case "promote":
			ruleName, _ := req.Params["rule_name"].(string)
			if err := deps.Policy.Promote(ruleName); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"success": true})
		case "dismiss", "investigate":
			writeJSON(w, http.StatusOK, map[string]any{"success": true})
		default:
			writeErrorString(w, http.StatusBadRequest, "unknown action")
		}
	})
}
