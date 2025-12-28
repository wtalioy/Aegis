package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aegis/pkg/events"
	"aegis/pkg/server"
	"aegis/pkg/storage"
	"aegis/pkg/utils"
)

func RegisterEventsHandlers(mux *http.ServeMux, app *server.App) {
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)

		if r.Method == "GET" && r.URL.Query().Get("stream") == "true" {
			// SSE stream (existing behavior)
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
		}

		// GET /api/events - List events with pagination
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")

			// Parse query parameters
			limit := 50
			if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
				if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
					limit = l
				}
			}

			eventType := r.URL.Query().Get("type")
			processName := r.URL.Query().Get("process")

			core := app.Core()
			if core == nil || core.Storage == nil {
				json.NewEncoder(w).Encode(map[string]any{
					"events": []any{},
					"total":  0,
					"page":   1,
				})
				return
			}

			// Get latest events
			eventList, err := core.Storage.Latest(limit)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Filter by type and process if specified
			filtered := make([]*storage.Event, 0)
			for _, ev := range eventList {
				if eventType != "" && string(ev.Type) != eventType {
					continue
				}
				if processName != "" {
					// Check process name in event data
					matched := false
					switch ev.Type {
					case events.EventTypeExec:
						switch v := ev.Data.(type) {
						case *events.ExecEvent:
							if utils.ExtractCString(v.Hdr.Comm[:]) == processName {
								matched = true
							}
						case events.ExecEvent:
							if utils.ExtractCString(v.Hdr.Comm[:]) == processName {
								matched = true
							}
						}
					case events.EventTypeFileOpen:
						switch v := ev.Data.(type) {
						case *events.FileOpenEvent:
							if utils.ExtractCString(v.Hdr.Comm[:]) == processName {
								matched = true
							}
						case events.FileOpenEvent:
							if utils.ExtractCString(v.Hdr.Comm[:]) == processName {
								matched = true
							}
						}
					case events.EventTypeConnect:
						switch v := ev.Data.(type) {
						case *events.ConnectEvent:
							if utils.ExtractCString(v.Hdr.Comm[:]) == processName {
								matched = true
							}
						case events.ConnectEvent:
							if utils.ExtractCString(v.Hdr.Comm[:]) == processName {
								matched = true
							}
						}
					}
					if !matched {
						continue
					}
				}
				filtered = append(filtered, ev)
			}

			// Convert to frontend format
			frontendEvents := convertEventsToFrontend(filtered)

			json.NewEncoder(w).Encode(map[string]any{
				"events": frontendEvents,
				"total":  len(frontendEvents),
				"page":   1,
			})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Phase 1: Get event by ID (using index)
	mux.HandleFunc("/api/events/", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/events/"), "/")
		if len(pathParts) == 0 {
			http.Error(w, "Invalid event ID", http.StatusBadRequest)
			return
		}

		if pathParts[0] == "range" {
			// GET /api/events/range - Time range query
			startStr := r.URL.Query().Get("start")
			endStr := r.URL.Query().Get("end")

			if startStr == "" || endStr == "" {
				http.Error(w, "start and end parameters required", http.StatusBadRequest)
				return
			}

			start, err := time.Parse(time.RFC3339, startStr)
			if err != nil {
				http.Error(w, "Invalid start time format (use RFC3339)", http.StatusBadRequest)
				return
			}

			end, err := time.Parse(time.RFC3339, endStr)
			if err != nil {
				http.Error(w, "Invalid end time format (use RFC3339)", http.StatusBadRequest)
				return
			}

			core := app.Core()
			if core == nil || core.Storage == nil {
				json.NewEncoder(w).Encode(map[string]any{
					"events": []any{},
				})
				return
			}

			eventList, err := core.Storage.Query(start, end)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Convert to frontend format
			frontendEvents := convertEventsToFrontend(eventList)

			json.NewEncoder(w).Encode(map[string]any{
				"events": frontendEvents,
			})
			return
		}

		// GET /api/events/{id} - Get event by ID
		eventID := pathParts[0]

		core := app.Core()
		if core == nil || core.Storage == nil {
			http.Error(w, "Storage not available", http.StatusServiceUnavailable)
			return
		}

		// Search through recent events to find matching ID
		// Event ID is generated as hash of timestamp + event data
		eventList, err := core.Storage.Latest(10000) // Search last 10k events
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var foundEvent *storage.Event
		for _, ev := range eventList {
			// Generate ID for this event
			id := generateEventID(ev)
			if id == eventID {
				foundEvent = ev
				break
			}
		}

		if foundEvent == nil {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}

		// Convert to frontend format
		frontendEvent := convertEventToFrontend(foundEvent)
		if frontendEvent == nil {
			http.Error(w, "Event format not supported", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(frontendEvent)
	})
}


func convertEventsToFrontend(events []*storage.Event) []any {
	frontendEvents := make([]any, 0, len(events))
	for _, ev := range events {
		if fe := convertEventToFrontend(ev); fe != nil {
			frontendEvents = append(frontendEvents, fe)
		}
	}
	return frontendEvents
}


func convertEventToFrontend(ev *storage.Event) any {
	switch ev.Type {
	case events.EventTypeExec:
		switch v := ev.Data.(type) {
		case *events.ExecEvent:
			return server.ExecToFrontend(*v)
		case events.ExecEvent:
			return server.ExecToFrontend(v)
		}
	case events.EventTypeFileOpen:
		switch v := ev.Data.(type) {
		case *events.FileOpenEvent:
			filename := utils.ExtractCString(v.Filename[:])
			return server.FileToFrontend(*v, filename)
		case events.FileOpenEvent:
			filename := utils.ExtractCString(v.Filename[:])
			return server.FileToFrontend(v, filename)
		}
	case events.EventTypeConnect:
		switch v := ev.Data.(type) {
		case *events.ConnectEvent:
			ip := utils.ExtractIP(v)
			addr := fmt.Sprintf("%s:%d", ip, v.Port)
			processName := utils.ExtractCString(v.Hdr.Comm[:])
			return server.ConnectToFrontend(*v, addr, processName)
		case events.ConnectEvent:
			ip := utils.ExtractIP(&v)
			addr := fmt.Sprintf("%s:%d", ip, v.Port)
			processName := utils.ExtractCString(v.Hdr.Comm[:])
			return server.ConnectToFrontend(v, addr, processName)
		}
	}
	return nil
}
