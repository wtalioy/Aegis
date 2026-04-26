package httpapi

import (
	"net/http"

	"aegis/internal/system"
)

type systemStatsResponse struct {
	ProcessCount  int     `json:"processCount"`
	WorkloadCount int     `json:"workloadCount"`
	EventsPerSec  float64 `json:"eventsPerSec"`
	AlertCount    int     `json:"alertCount"`
	ProbeStatus   string  `json:"probeStatus"`
	ProbeError    string  `json:"probeError,omitempty"`
}

func registerSystemRoutes(mux *http.ServeMux, deps Dependencies) {
	registerAliases(mux, []string{"/api/v1/system/stats"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		processCount := 0
		if deps.Telemetry != nil && deps.Telemetry.ProcessTree() != nil {
			processCount = deps.Telemetry.ProcessTree().Size()
		}
		execRate, fileRate, connectRate := deps.Stats.Rates()
		probe := system.ProbeStatus{Status: system.ProbeStatusStarting}
		if deps.ProbeStatus != nil {
			probe = deps.ProbeStatus.ProbeStatus()
			if probe.Status == "" {
				probe.Status = system.ProbeStatusStarting
			}
		}
		writeJSON(w, http.StatusOK, systemStatsResponse{
			ProcessCount:  processCount,
			WorkloadCount: deps.Stats.WorkloadCount(),
			EventsPerSec:  float64(execRate + fileRate + connectRate),
			AlertCount:    int(deps.Stats.TotalAlertCount()),
			ProbeStatus:   probe.Status,
			ProbeError:    probe.Error,
		})
	})

	registerAliases(mux, []string{"/api/v1/system/alerts"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		writeJSON(w, http.StatusOK, deps.Stats.Alerts())
	})

	registerAliases(mux, []string{"/api/v1/system/alerts/stream"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		subscription := deps.AlertStream.Subscribe(100)
		streamSSE(w, r, subscription.C, subscription.Cancel, nil)
	})

	registerAliases(mux, []string{"/api/v1/system/settings"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, deps.Settings.Get())
		case http.MethodPut:
			var cfg system.Settings
			if err := decodeJSON(r, &cfg); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			result, err := deps.Settings.Update(cfg)
			if err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusOK, result)
		case http.MethodOptions:
			allowJSONOptions(w, http.MethodGet, http.MethodPut)
		default:
			methodNotAllowed(w)
		}
	})
}
