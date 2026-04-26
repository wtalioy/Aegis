package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"aegis/internal/analysis/sentinel"
	"aegis/internal/analysis/types"
)

type sentinelInsightsResponse struct {
	Insights []*sentinel.Insight `json:"insights"`
	Total    int                 `json:"total"`
}

func registerSentinelRoutes(mux *http.ServeMux, deps Dependencies) {
	registerAliases(mux, []string{"/api/v1/sentinel/stream"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		if deps.Analysis == nil || deps.Analysis.Sentinel() == nil {
			writeErrorString(w, http.StatusServiceUnavailable, "Sentinel not available")
			return
		}
		subscription := deps.Analysis.Sentinel().Subscribe()
		streamSSE(w, r, subscription.C, subscription.Cancel, nil)
	})

	registerAliases(mux, []string{"/api/v1/sentinel/insights"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		if deps.Analysis == nil || deps.Analysis.Sentinel() == nil {
			writeJSON(w, http.StatusOK, sentinelInsightsResponse{Insights: []*sentinel.Insight{}, Total: 0})
			return
		}
		limit := 50
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		insights := deps.Analysis.Sentinel().GetInsights(limit)
		writeJSON(w, http.StatusOK, sentinelInsightsResponse{Insights: insights, Total: len(insights)})
	})

	registerAliases(mux, []string{"/api/v1/sentinel/ask"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodPost) {
			return
		}
		if deps.Analysis == nil {
			writeErrorString(w, http.StatusServiceUnavailable, "AI service not available")
			return
		}
		var req types.AskInsightRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		ctx, cancel := withTimeout(r, 90*time.Second)
		defer cancel()
		result, err := deps.Analysis.AskAboutInsight(ctx, &req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
}
