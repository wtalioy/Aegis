package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aegis/pkg/events"
	"aegis/pkg/server"
	"aegis/pkg/storage"
	"aegis/pkg/utils"
)

func RegisterQueryHandlers(mux *http.ServeMux, app *server.App) {
	// Phase 3: Event Query
	mux.HandleFunc("/api/query", func(w http.ResponseWriter, r *http.Request) {
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
			Page      int    `json:"page"`
			Limit     int    `json:"limit"`
			SortBy    string `json:"sort_by"`
			SortOrder string `json:"sort_order"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		core := app.Core()
		if core == nil || core.Storage == nil {
			http.Error(w, "Storage not available", http.StatusServiceUnavailable)
			return
		}

		// Build storage filter
		filter := storage.Filter{
			PIDs:      req.Filter.PIDs,
			CgroupIDs: req.Filter.CgroupIDs,
			Processes: req.Filter.Processes,
		}

		// Convert type strings to EventType
		for _, t := range req.Filter.Types {
			switch strings.ToLower(t) {
			case "exec":
				filter.Types = append(filter.Types, events.EventTypeExec)
			case "file", "fileopen":
				filter.Types = append(filter.Types, events.EventTypeFileOpen)
			case "connect", "network":
				filter.Types = append(filter.Types, events.EventTypeConnect)
			}
		}

		// Query events
		var allEvents []*storage.Event
		var err error

		if req.Filter.TimeWindow.Start != "" && req.Filter.TimeWindow.End != "" {
			startTime, err1 := time.Parse(time.RFC3339, req.Filter.TimeWindow.Start)
			endTime, err2 := time.Parse(time.RFC3339, req.Filter.TimeWindow.End)
			if err1 == nil && err2 == nil {
				allEvents, err = core.Storage.Query(startTime, endTime)
			} else {
				allEvents, err = core.Storage.Latest(10000)
			}
		} else {
			allEvents, err = core.Storage.Latest(10000)
		}

		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to query events: %v", err), http.StatusInternalServerError)
			return
		}

		// Apply filters
		filteredEvents := make([]*storage.Event, 0)
		for _, event := range allEvents {
			if matchesFilter(event, &filter) {
				filteredEvents = append(filteredEvents, event)
			}
		}

		// Paginate
		total := len(filteredEvents)
		page := req.Page
		if page < 1 {
			page = 1
		}
		limit := req.Limit
		if limit < 1 {
			limit = 50
		}
		if limit > 1000 {
			limit = 1000
		}

		start := (page - 1) * limit
		end := start + limit
		if end > total {
			end = total
		}

		var displayedEvents []*storage.Event
		if start < total {
			displayedEvents = filteredEvents[start:end]
		} else {
			displayedEvents = make([]*storage.Event, 0)
		}

		totalPages := (total + limit - 1) / limit

		// Calculate type counts
		typeCounts := struct {
			Exec    int `json:"exec"`
			File    int `json:"file"`
			Connect int `json:"connect"`
		}{}
		for _, ev := range filteredEvents {
			if ev == nil {
				continue
			}
			switch ev.Type {
			case events.EventTypeExec:
				typeCounts.Exec++
			case events.EventTypeFileOpen:
				typeCounts.File++
			case events.EventTypeConnect:
				typeCounts.Connect++
			}
		}

		// Convert storage events to frontend events
		frontendEvents := make([]any, 0, len(displayedEvents))
		for _, ev := range displayedEvents {
			if ev == nil {
				continue
			}
			switch ev.Type {
			case events.EventTypeExec:
				var execEv *events.ExecEvent
				if ptr, ok := ev.Data.(*events.ExecEvent); ok {
					execEv = ptr
				} else if val, ok := ev.Data.(events.ExecEvent); ok {
					execEv = &val
				}
				if execEv != nil {
					frontendEvents = append(frontendEvents, server.ExecToFrontend(*execEv))
				}
			case events.EventTypeFileOpen:
				var fileEv *events.FileOpenEvent
				if ptr, ok := ev.Data.(*events.FileOpenEvent); ok {
					fileEv = ptr
				} else if val, ok := ev.Data.(events.FileOpenEvent); ok {
					fileEv = &val
				}
				if fileEv != nil {
					filename := utils.ExtractCString(fileEv.Filename[:])
					frontendEvents = append(frontendEvents, server.FileToFrontend(*fileEv, filename))
				}
			case events.EventTypeConnect:
				var connEv *events.ConnectEvent
				if ptr, ok := ev.Data.(*events.ConnectEvent); ok {
					connEv = ptr
				} else if val, ok := ev.Data.(events.ConnectEvent); ok {
					connEv = &val
				}
				if connEv != nil {
					ip := utils.ExtractIP(connEv)
					addr := fmt.Sprintf("%s:%d", ip, connEv.Port)
					processName := utils.ExtractCString(connEv.Hdr.Comm[:])
					frontendEvents = append(frontendEvents, server.ConnectToFrontend(*connEv, addr, processName))
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"events":      frontendEvents,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
			"type_counts": typeCounts,
		})
	})
}
