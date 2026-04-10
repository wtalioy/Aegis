package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aegis/internal/telemetry"
)

func registerEventRoutes(mux *http.ServeMux, deps Dependencies) {
	eventStreamHandler := func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		streamSSE(w, r, deps.EventStream.Subscribe(100), func(event telemetry.Event) any {
			return eventToDTO(event)
		})
	}

	registerAliases(mux, []string{"/api/v1/events/stream"}, eventStreamHandler)

	registerAliases(mux, []string{"/api/v1/events"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method == http.MethodGet && r.URL.Query().Get("stream") == "true" {
			eventStreamHandler(w, r)
			return
		}
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}

		limit := 50
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		filter := telemetry.Filter{}
		if eventType := strings.TrimSpace(r.URL.Query().Get("type")); eventType != "" {
			if parsed := parseEventType(eventType); parsed != "" {
				filter.Types = []telemetry.EventType{parsed}
			}
		}
		if process := strings.TrimSpace(r.URL.Query().Get("process")); process != "" {
			filter.Processes = []string{process}
		}

		result := deps.Telemetry.Query(telemetry.Query{
			Filter: filter,
			Page:   1,
			Limit:  limit,
		})
		writeJSON(w, http.StatusOK, map[string]any{
			"events":      mapEvents(result.Events),
			"total":       result.Total,
			"page":        result.Page,
			"limit":       result.Limit,
			"totalPages":  result.TotalPages,
			"type_counts": result.TypeCounts,
		})
	})

	registerAliases(mux, []string{"/api/v1/events/query"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			return
		}
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}

		var req struct {
			Filter struct {
				Types      []string `json:"types"`
				Processes  []string `json:"processes"`
				PIDs       []uint32 `json:"pids"`
				CgroupIDs  []uint64 `json:"cgroup_ids"`
				TimeWindow struct {
					Start string `json:"start"`
					End   string `json:"end"`
				} `json:"time_window"`
			} `json:"filter"`
			Page  int `json:"page"`
			Limit int `json:"limit"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		filter := telemetry.Filter{
			Processes: req.Filter.Processes,
			PIDs:      req.Filter.PIDs,
			CgroupIDs: req.Filter.CgroupIDs,
		}
		for _, rawType := range req.Filter.Types {
			if eventType := parseEventType(rawType); eventType != "" {
				filter.Types = append(filter.Types, eventType)
			}
		}
		if req.Filter.TimeWindow.Start != "" {
			if ts, err := time.Parse(time.RFC3339, req.Filter.TimeWindow.Start); err == nil {
				filter.Start = &ts
			}
		}
		if req.Filter.TimeWindow.End != "" {
			if ts, err := time.Parse(time.RFC3339, req.Filter.TimeWindow.End); err == nil {
				filter.End = &ts
			}
		}

		result := deps.Telemetry.Query(telemetry.Query{
			Filter: filter,
			Page:   req.Page,
			Limit:  req.Limit,
		})
		writeJSON(w, http.StatusOK, map[string]any{
			"events":      mapEvents(result.Events),
			"total":       result.Total,
			"page":        result.Page,
			"limit":       result.Limit,
			"totalPages":  result.TotalPages,
			"type_counts": result.TypeCounts,
		})
	})

	registerAliasesWithPrefix(mux, []string{"/api/v1/events/"}, func(w http.ResponseWriter, r *http.Request, id string) {
		setCORS(w)
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		if id == "" || id == "stream" || id == "query" {
			http.NotFound(w, r)
			return
		}
		record, ok := deps.Telemetry.Get(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, eventToDTO(record.Event))
	})
}

func mapEvents(events []telemetry.Event) []map[string]any {
	out := make([]map[string]any, 0, len(events))
	for _, event := range events {
		out = append(out, eventToDTO(event))
	}
	return out
}

func eventToDTO(event telemetry.Event) map[string]any {
	dto := map[string]any{
		"id":          event.ID,
		"type":        string(event.Type),
		"timestamp":   event.Timestamp.UnixMilli(),
		"pid":         event.PID,
		"cgroupId":    strconv.FormatUint(event.CgroupID, 10),
		"blocked":     event.Blocked,
		"processName": event.ProcessName,
	}

	switch event.Type {
	case telemetry.EventTypeExec:
		dto["comm"] = event.ProcessName
		dto["ppid"] = event.PPID
		dto["parentComm"] = event.ParentName
		dto["commandLine"] = event.CommandLine
	case telemetry.EventTypeFile:
		dto["filename"] = event.Filename
		dto["flags"] = event.Flags
		dto["ino"] = event.Ino
		dto["dev"] = event.Dev
	case telemetry.EventTypeConnect:
		dto["family"] = event.Family
		dto["port"] = event.Port
		dto["addr"] = event.Address
	}
	return dto
}

func parseEventType(raw string) telemetry.EventType {
	switch strings.ToLower(raw) {
	case "exec":
		return telemetry.EventTypeExec
	case "file", "fileopen":
		return telemetry.EventTypeFile
	case "connect", "network":
		return telemetry.EventTypeConnect
	default:
		return ""
	}
}
