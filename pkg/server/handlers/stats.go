package handlers

import (
	"encoding/json"
	"net/http"

	"aegis/pkg/server"
)

func RegisterStatsHandlers(mux *http.ServeMux, app *server.App) {
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")
		s := app.GetSystemStats()
		json.NewEncoder(w).Encode(map[string]any{
			"processCount":  s.ProcessCount,
			"workloadCount": s.WorkloadCount,
			"eventsPerSec":  s.EventsPerSec,
			"alertCount":    s.AlertCount,
			"probeStatus":   s.ProbeStatus,
		})
	})

	mux.HandleFunc("/api/stats/rates", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		execRate, fileRate, connectRate := app.Stats().Rates()
		json.NewEncoder(w).Encode(map[string]float64{
			"execRate":    float64(execRate),
			"fileRate":    float64(fileRate),
			"connectRate": float64(connectRate),
		})
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
			json.NewEncoder(w).Encode(map[string]any{
				"id":          a.ID,
				"timestamp":   a.Timestamp,
				"severity":    a.Severity,
				"ruleName":    a.RuleName,
				"description": a.Description,
				"pid":         a.PID,
				"processName": a.ProcessName,
				"cgroupId":    a.CgroupID,
				"action":      a.Action,
				"blocked":     a.Blocked,
			})
		}
		w.Write([]byte("]"))
	})
}
