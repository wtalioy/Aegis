package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"aegis/internal/telemetry"
)

type eventQueryRequest struct {
	Filter eventQueryFilter `json:"filter"`
	Page   int              `json:"page"`
	Limit  int              `json:"limit"`
}

type eventQueryFilter struct {
	Types     []string `json:"types"`
	Processes []string `json:"processes"`
	PIDs      []uint32 `json:"pids"`
	CgroupIDs []uint64 `json:"cgroupIds"`
	TimeWindow struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"timeWindow"`
}

type eventDTO struct {
	ID          string              `json:"id"`
	Type        telemetry.EventType `json:"type"`
	Timestamp   int64               `json:"timestamp"`
	PID         uint32              `json:"pid"`
	PPID        uint32              `json:"ppid,omitempty"`
	CgroupID    string              `json:"cgroupId"`
	ProcessName string              `json:"processName"`
	ParentComm  string              `json:"parentComm,omitempty"`
	CommandLine string              `json:"commandLine,omitempty"`
	Filename    string              `json:"filename,omitempty"`
	Flags       uint32              `json:"flags,omitempty"`
	Ino         uint64              `json:"ino,omitempty"`
	Dev         uint64              `json:"dev,omitempty"`
	Family      uint16              `json:"family,omitempty"`
	Port        uint16              `json:"port,omitempty"`
	Addr        string              `json:"addr,omitempty"`
	Blocked     bool                `json:"blocked"`
}

type eventPageResponse struct {
	Events     []eventDTO           `json:"events"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"totalPages"`
	TypeCounts telemetry.TypeCounts `json:"typeCounts"`
}

func registerEventRoutes(mux *http.ServeMux, deps Dependencies) {
	eventStreamHandler := func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		subscription := deps.EventStream.Subscribe(100)
		streamMappedSSE(w, r, subscription.C, subscription.Cancel, func(event telemetry.Event) eventDTO {
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
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		result := deps.Telemetry.Query(queryFromListRequest(readEventListRequest(r)))
		writeJSON(w, http.StatusOK, toEventPageResponse(result))
	})

	registerAliases(mux, []string{"/api/v1/events/query"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodPost) {
			return
		}

		var req eventQueryRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		result := deps.Telemetry.Query(queryFromRequest(req))
		writeJSON(w, http.StatusOK, toEventPageResponse(result))
	})

	registerAliasesWithPrefix(mux, []string{"/api/v1/events/"}, func(w http.ResponseWriter, r *http.Request, id string) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodGet) {
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

func readEventListRequest(r *http.Request) eventQueryRequest {
	req := eventQueryRequest{
		Page:  1,
		Limit: 50,
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			req.Limit = parsed
		}
	}
	if eventType := strings.TrimSpace(r.URL.Query().Get("type")); eventType != "" {
		req.Filter.Types = []string{eventType}
	}
	if process := strings.TrimSpace(r.URL.Query().Get("process")); process != "" {
		req.Filter.Processes = []string{process}
	}
	return req
}

func mapEvents(events []telemetry.Event) []eventDTO {
	out := make([]eventDTO, 0, len(events))
	for _, event := range events {
		out = append(out, eventToDTO(event))
	}
	return out
}

func eventToDTO(event telemetry.Event) eventDTO {
	dto := eventDTO{
		ID:          event.ID,
		Type:        event.Type,
		Timestamp:   event.Timestamp.UnixMilli(),
		PID:         event.PID,
		CgroupID:    strconv.FormatUint(event.CgroupID, 10),
		ProcessName: event.ProcessName,
		Blocked:     event.Blocked,
	}
	switch event.Type {
	case telemetry.EventTypeExec:
		dto.PPID = event.PPID
		dto.ParentComm = event.ParentName
		dto.CommandLine = event.CommandLine
		dto.Filename = event.Filename
	case telemetry.EventTypeFile:
		dto.Filename = event.Filename
		dto.Flags = event.Flags
		dto.Ino = event.Ino
		dto.Dev = event.Dev
	case telemetry.EventTypeConnect:
		dto.Family = event.Family
		dto.Port = event.Port
		dto.Addr = event.Address
	}
	return dto
}

func toEventPageResponse(result telemetry.PageResult) eventPageResponse {
	return eventPageResponse{
		Events:     mapEvents(result.Events),
		Total:      result.Total,
		Page:       result.Page,
		Limit:      result.Limit,
		TotalPages: result.TotalPages,
		TypeCounts: result.TypeCounts,
	}
}

func queryFromListRequest(req eventQueryRequest) telemetry.Query {
	return queryFromRequest(req)
}

func queryFromRequest(req eventQueryRequest) telemetry.Query {
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
	filter.Start = parseRFC3339Time(req.Filter.TimeWindow.Start)
	filter.End = parseRFC3339Time(req.Filter.TimeWindow.End)

	return telemetry.Query{
		Filter: filter,
		Page:   req.Page,
		Limit:  req.Limit,
	}
}

func parseRFC3339Time(raw string) *time.Time {
	if raw == "" {
		return nil
	}
	ts, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return &ts
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
