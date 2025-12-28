package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aegis/pkg/events"
	"aegis/pkg/storage"
)

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

// generateEventID generates a unique ID for an event based on its content and timestamp
func generateEventID(event *storage.Event) string {
	// Create a hash from timestamp and event data
	h := sha256.New()
	h.Write([]byte(event.Timestamp.Format(time.RFC3339Nano)))

	// Add event type
	fmt.Fprintf(h, "%d", int(event.Type))

	// Add event-specific data
	switch ev := event.Data.(type) {
	case *events.ExecEvent:
		h.Write(ev.Hdr.Comm[:])
		fmt.Fprintf(h, "%d", ev.Hdr.PID)
	case *events.FileOpenEvent:
		h.Write(ev.Filename[:])
		fmt.Fprintf(h, "%d", ev.Hdr.PID)
	case *events.ConnectEvent:
		fmt.Fprintf(h, "%d:%d", ev.Port, ev.Hdr.PID)
	}

	return hex.EncodeToString(h.Sum(nil))[:16] // Use first 16 chars as ID
}

// matchesFilter checks if an event matches the filter criteria
func matchesFilter(event *storage.Event, filter *storage.Filter) bool {
	if filter == nil {
		return true
	}

	// Check type
	if len(filter.Types) > 0 {
		matched := false
		for _, t := range filter.Types {
			if event.Type == t {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check PID
	if len(filter.PIDs) > 0 {
		matched := false
		var pid uint32
		switch ev := event.Data.(type) {
		case *events.ExecEvent:
			pid = ev.Hdr.PID
		case *events.FileOpenEvent:
			pid = ev.Hdr.PID
		case *events.ConnectEvent:
			pid = ev.Hdr.PID
		}
		for _, p := range filter.PIDs {
			if pid == p {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check CgroupID
	if len(filter.CgroupIDs) > 0 {
		matched := false
		var cgroupID uint64
		switch ev := event.Data.(type) {
		case *events.ExecEvent:
			cgroupID = ev.Hdr.CgroupID
		case *events.FileOpenEvent:
			cgroupID = ev.Hdr.CgroupID
		case *events.ConnectEvent:
			cgroupID = ev.Hdr.CgroupID
		}
		for _, c := range filter.CgroupIDs {
			if cgroupID == c {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check process name
	if len(filter.Processes) > 0 {
		matched := false
		var processName string
		switch ev := event.Data.(type) {
		case *events.ExecEvent:
			processName = strings.TrimRight(string(ev.Hdr.Comm[:]), "\x00")
		case *events.FileOpenEvent:
			processName = strings.TrimRight(string(ev.Hdr.Comm[:]), "\x00")
		case *events.ConnectEvent:
			processName = strings.TrimRight(string(ev.Hdr.Comm[:]), "\x00")
		}
		for _, p := range filter.Processes {
			if strings.Contains(processName, p) || strings.Contains(p, processName) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

