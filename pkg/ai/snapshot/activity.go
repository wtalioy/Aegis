package snapshot

import (
	"fmt"
	"time"

	"aegis/pkg/apimodel"
	"aegis/pkg/events"
	"aegis/pkg/storage"
	"aegis/pkg/utils"
)

// recentActivity holds the three types of recent activity summaries
type recentActivity struct {
	Processes   []ProcessActivity
	Connections []ConnectionActivity
	Files       []FileActivity
}

// buildRecentActivity queries storage and builds all recent activity summaries
func (s *Snapshot) buildRecentActivity() ([]apimodel.ExecEvent, recentActivity) {
	if s.store == nil {
		return nil, recentActivity{}
	}

	now := time.Now()
	windowStart := now.Add(-RecentEventWindow)

	execEvents := s.queryEventsByType(events.EventTypeExec, windowStart, now)
	connectEvents := s.queryEventsByType(events.EventTypeConnect, windowStart, now)
	fileEvents := s.queryEventsByType(events.EventTypeFileOpen, windowStart, now)

	// Fallback: if queries return nothing but the store has events, it usually means
	// the event timestamps are not aligned with time.Now() (kernel->wall conversion / clock skew).
	if len(execEvents) == 0 && len(connectEvents) == 0 && len(fileEvents) == 0 {
		if latest, err := s.store.Latest(1); err == nil && len(latest) == 1 && latest[0] != nil {
			latestTs := latest[0].Timestamp
			fallbackStart := latestTs.Add(-RecentEventWindow)
			fallbackEnd := latestTs.Add(10 * time.Second)
			execEvents = s.queryEventsByType(events.EventTypeExec, fallbackStart, fallbackEnd)
			connectEvents = s.queryEventsByType(events.EventTypeConnect, fallbackStart, fallbackEnd)
			fileEvents = s.queryEventsByType(events.EventTypeFileOpen, fallbackStart, fallbackEnd)
		}
	}

	procSummaries := buildProcessActivityFromStorage(execEvents)
	connSummaries := buildConnectionActivityFromStorage(connectEvents)
	fileSummaries := buildFileActivityFromStorage(fileEvents)

	return convertStorageExecEvents(execEvents), recentActivity{
		Processes:   procSummaries,
		Connections: connSummaries,
		Files:       fileSummaries,
	}
}

// queryEventsByType queries events of a specific type from storage
func (s *Snapshot) queryEventsByType(eventType events.EventType, start, end time.Time) []*storage.Event {
	if manager, ok := s.store.(*storage.Manager); ok {
		return manager.QueryByTypeWithTimeRange(eventType, start, end, MaxEventsPerType)
	}

	// Fallback for other EventStore implementations
	allEvents, err := s.store.Query(start, end)
	if err != nil {
		return nil
	}

	result := make([]*storage.Event, 0, MaxEventsPerType)
	for _, ev := range allEvents {
		if ev != nil && ev.Type == eventType && len(result) < MaxEventsPerType {
			result = append(result, ev)
		}
	}
	return result
}

func convertStorageExecEvents(storageEvents []*storage.Event) []apimodel.ExecEvent {
	execs := make([]apimodel.ExecEvent, 0, len(storageEvents))
	for _, sev := range storageEvents {
		switch v := sev.Data.(type) {
		case *events.ExecEvent:
			execEv := v
			// Use storage timestamp rather than header time for recency.
			execs = append(execs, apimodel.ExecEvent{
				Type:        "exec",
				Timestamp:   sev.Timestamp.UnixMilli(),
				PID:         execEv.Hdr.PID,
				PPID:        execEv.PPID,
				CgroupID:    fmt.Sprintf("%d", execEv.Hdr.CgroupID),
				Comm:        utils.ExtractCString(execEv.Hdr.Comm[:]),
				ParentComm:  utils.ExtractCString(execEv.PComm[:]),
				CommandLine: utils.ExtractCString(execEv.CommandLine[:]),
				Blocked:     execEv.Hdr.Blocked == 1,
			})
		case events.ExecEvent:
			execEv := &v
			execs = append(execs, apimodel.ExecEvent{
				Type:        "exec",
				Timestamp:   sev.Timestamp.UnixMilli(),
				PID:         execEv.Hdr.PID,
				PPID:        execEv.PPID,
				CgroupID:    fmt.Sprintf("%d", execEv.Hdr.CgroupID),
				Comm:        utils.ExtractCString(execEv.Hdr.Comm[:]),
				ParentComm:  utils.ExtractCString(execEv.PComm[:]),
				CommandLine: utils.ExtractCString(execEv.CommandLine[:]),
				Blocked:     execEv.Hdr.Blocked == 1,
			})
		default:
			// ignore
		}
	}
	return execs
}

func buildProcessActivityFromStorage(storageEvents []*storage.Event) []ProcessActivity {
	groups := make(map[string]*ProcessActivity)

	for _, sev := range storageEvents {
		var execEv *events.ExecEvent
		switch v := sev.Data.(type) {
		case *events.ExecEvent:
			execEv = v
		case events.ExecEvent:
			execEv = &v
		default:
			continue
		}

		comm := utils.ExtractCString(execEv.Hdr.Comm[:])
		parentComm := utils.ExtractCString(execEv.PComm[:])

		key := comm + "|" + parentComm
		blocked := execEv.Hdr.Blocked == 1

		if existing, ok := groups[key]; ok {
			existing.Count++
			existing.Blocked = existing.Blocked || blocked
		} else {
			groups[key] = &ProcessActivity{
				Comm:       comm,
				ParentComm: parentComm,
				Count:      1,
				Blocked:    blocked,
			}
		}
	}

	return finalizeGroup(groups, MaxAlertSummaries, func(a, b ProcessActivity) bool {
		return compareByBlockedThenCount(a.Blocked, a.Count, b.Blocked, b.Count)
	})
}

func buildConnectionActivityFromStorage(storageEvents []*storage.Event) []ConnectionActivity {
	groups := make(map[string]*ConnectionActivity)

	for _, sev := range storageEvents {
		connEv, ok := sev.Data.(*events.ConnectEvent)
		if !ok {
			continue
		}

		addr := utils.ExtractIP(connEv)
		if addr == "" {
			continue
		}
		if connEv.Port != 0 {
			addr = fmt.Sprintf("%s:%d", addr, connEv.Port)
		}

		blocked := connEv.Hdr.Blocked == 1
		if existing, ok := groups[addr]; ok {
			existing.Count++
			existing.Blocked = existing.Blocked || blocked
		} else {
			groups[addr] = &ConnectionActivity{
				Destination: addr,
				Count:       1,
				Blocked:     blocked,
			}
		}
	}

	return finalizeGroup(groups, MaxActivitySummaries, func(a, b ConnectionActivity) bool {
		return compareByBlockedThenCount(a.Blocked, a.Count, b.Blocked, b.Count)
	})
}

func buildFileActivityFromStorage(storageEvents []*storage.Event) []FileActivity {
	groups := make(map[string]*FileActivity)

	for _, sev := range storageEvents {
		fileEv, ok := sev.Data.(*events.FileOpenEvent)
		if !ok {
			continue
		}

		filename := utils.ExtractCString(fileEv.Filename[:])
		if filename == "" {
			continue
		}

		path := simplifyFilePath(filename)
		blocked := fileEv.Hdr.Blocked == 1

		if existing, ok := groups[path]; ok {
			existing.Count++
			existing.Blocked = existing.Blocked || blocked
		} else {
			groups[path] = &FileActivity{
				Path:    path,
				Count:   1,
				Blocked: blocked,
			}
		}
	}

	return finalizeGroup(groups, MaxActivitySummaries, func(a, b FileActivity) bool {
		return compareByBlockedThenCount(a.Blocked, a.Count, b.Blocked, b.Count)
	})
}
