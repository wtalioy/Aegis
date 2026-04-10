package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aegis/internal/shared/stream"
)

func registerAliases(mux *http.ServeMux, paths []string, handler http.HandlerFunc) {
	for _, path := range paths {
		if handler == nil {
			continue
		}
		mux.HandleFunc(path, handler)
	}
}

func registerAliasesWithPrefix(mux *http.ServeMux, prefixes []string, handler func(http.ResponseWriter, *http.Request, string)) {
	for _, prefix := range prefixes {
		localPrefix := prefix
		mux.HandleFunc(localPrefix, func(w http.ResponseWriter, r *http.Request) {
			suffix := strings.TrimPrefix(r.URL.Path, localPrefix)
			handler(w, r, strings.Trim(suffix, "/"))
		})
	}
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeErrorString(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

func streamSSE[T any](w http.ResponseWriter, r *http.Request, subscription stream.Subscription[T], mapper func(T) any) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
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
		case item, ok := <-subscription.C:
			if !ok {
				return
			}
			payload := item
			if mapper != nil {
				if mapped, ok := mapper(item).(T); ok {
					payload = mapped
				} else {
					data, _ := json.Marshal(mapper(item))
					fmt.Fprintf(w, "data: %s\n\n", data)
					flusher.Flush()
					continue
				}
			}
			data, _ := json.Marshal(payload)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprintf(w, "data: {\"type\":\"heartbeat\"}\n\n")
			flusher.Flush()
		}
	}
}
