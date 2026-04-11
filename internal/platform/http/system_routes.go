package httpapi

import (
	"encoding/json"
	"net/http"

	"aegis/internal/system"
)

func registerSystemRoutes(mux *http.ServeMux, deps Dependencies) {
	registerAliases(mux, []string{"/api/v1/system/stats"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
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
		writeJSON(w, http.StatusOK, map[string]any{
			"processCount":  processCount,
			"workloadCount": deps.Stats.WorkloadCount(),
			"eventsPerSec":  float64(execRate + fileRate + connectRate),
			"alertCount":    int(deps.Stats.TotalAlertCount()),
			"probeStatus":   probe.Status,
			"probeError":    probe.Error,
		})
	})

	registerAliases(mux, []string{"/api/v1/system/alerts"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, deps.Stats.Alerts())
	})

	registerAliases(mux, []string{"/api/v1/system/alerts/stream"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		streamSSE(w, r, deps.AlertStream.Subscribe(100), func(alert system.Alert) any {
			return alert
		})
	})

	registerAliases(mux, []string{"/api/v1/system/settings"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, deps.Settings.Get())
		case http.MethodPut:
			var cfg system.Settings
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
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
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		default:
			methodNotAllowed(w)
		}
	})
}
