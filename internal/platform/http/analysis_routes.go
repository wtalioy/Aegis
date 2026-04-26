package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"aegis/internal/analysis/types"
)

func registerAnalysisRoutes(mux *http.ServeMux, deps Dependencies) {
	registerAliases(mux, []string{"/api/v1/analysis/status"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		if deps.Analysis == nil {
			writeJSON(w, http.StatusOK, types.StatusDTO{Status: "unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, deps.Analysis.Status())
	})

	registerAliases(mux, []string{"/api/v1/analysis/diagnose"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		if deps.Analysis == nil {
			writeErrorString(w, http.StatusServiceUnavailable, "AI service not available")
			return
		}
		ctx, cancel := withTimeout(r, 90*time.Second)
		defer cancel()
		result, err := deps.Analysis.Diagnose(ctx)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	registerAliases(mux, []string{"/api/v1/analysis/chat"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodPost, http.MethodPost) {
			return
		}
		if deps.Analysis == nil {
			writeErrorString(w, http.StatusServiceUnavailable, "AI service not available")
			return
		}
		var req struct {
			Message   string `json:"message"`
			SessionID string `json:"sessionId"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if req.SessionID == "" {
			req.SessionID = generateSessionID()
		}
		ctx, cancel := withTimeout(r, 90*time.Second)
		defer cancel()
		result, err := deps.Analysis.Chat(ctx, req.SessionID, req.Message)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	registerAliases(mux, []string{"/api/v1/analysis/chat/stream"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodPost, http.MethodPost) {
			return
		}
		if deps.Analysis == nil {
			writeErrorString(w, http.StatusServiceUnavailable, "AI service not available")
			return
		}
		var req struct {
			Message   string `json:"message"`
			SessionID string `json:"sessionId"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if req.SessionID == "" {
			req.SessionID = generateSessionID()
		}

		ctx, cancel := withTimeout(r, 120*time.Second)
		defer cancel()
		tokenChan, err := deps.Analysis.ChatStream(ctx, req.SessionID, req.Message)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, err)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		flusher, ok := w.(http.Flusher)
		if !ok {
			writeErrorString(w, http.StatusInternalServerError, "streaming not supported")
			return
		}
		for token := range tokenChan {
			data, _ := json.Marshal(token)
			_, _ = w.Write([]byte("data: "))
			_, _ = w.Write(data)
			_, _ = w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	})

	registerAliases(mux, []string{"/api/v1/analysis/chat/history"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		if deps.Analysis == nil {
			writeJSON(w, http.StatusOK, []types.Message{})
			return
		}
		sessionID := r.URL.Query().Get("sessionId")
		writeJSON(w, http.StatusOK, deps.Analysis.ChatHistory(sessionID))
	})

	registerAliases(mux, []string{"/api/v1/analysis/chat/clear"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodPost) {
			return
		}
		if deps.Analysis != nil {
			var req struct {
				SessionID string `json:"sessionId"`
			}
			_ = decodeJSON(r, &req)
			deps.Analysis.ClearChat(req.SessionID)
		}
		writeJSON(w, http.StatusOK, successResponse{Success: true})
	})

	registerAliases(mux, []string{"/api/v1/analysis/generate-rule"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodPost) {
			return
		}
		if deps.Analysis == nil {
			writeErrorString(w, http.StatusServiceUnavailable, "AI service not available")
			return
		}
		var req types.RuleGenRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		ctx, cancel := withTimeout(r, 90*time.Second)
		defer cancel()
		result, err := deps.Analysis.GenerateRule(ctx, &req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	registerAliases(mux, []string{"/api/v1/analysis/explain"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodPost) {
			return
		}
		if deps.Analysis == nil {
			writeErrorString(w, http.StatusServiceUnavailable, "AI service not available")
			return
		}
		var req types.ExplainRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if req.EventID == "" {
			writeErrorString(w, http.StatusBadRequest, "eventId is required")
			return
		}
		ctx, cancel := withTimeout(r, 90*time.Second)
		defer cancel()
		result, err := deps.Analysis.ExplainEvent(ctx, &req, req.EventID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	registerAliases(mux, []string{"/api/v1/analysis/analyze"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodPost) {
			return
		}
		if deps.Analysis == nil {
			writeErrorString(w, http.StatusServiceUnavailable, "AI service not available")
			return
		}
		var req types.AnalyzeRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		ctx, cancel := withTimeout(r, 90*time.Second)
		defer cancel()
		result, err := deps.Analysis.Analyze(ctx, &req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
}
