package storage

import (
	"sync"

	"aegis/pkg/events"
)

type Indexer struct {
	pidIndex     map[uint32][]*Event
	cgroupIndex  map[uint64][]*Event
	typeIndex    map[events.EventType][]*Event
	processIndex map[string][]*Event
	mu           sync.RWMutex
	maxIndexSize int // Maximum events per index entry to prevent memory bloat
}

func NewIndexer(maxIndexSize int) *Indexer {
	if maxIndexSize <= 0 {
		maxIndexSize = 1000 // Default: keep last 1000 events per index entry
	}
	return &Indexer{
		pidIndex:     make(map[uint32][]*Event),
		cgroupIndex:  make(map[uint64][]*Event),
		typeIndex:    make(map[events.EventType][]*Event),
		processIndex: make(map[string][]*Event),
		maxIndexSize: maxIndexSize,
	}
}

func (idx *Indexer) IndexEvent(event *Event) {
	if event == nil {
		return
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Index by event type
	eventList := idx.typeIndex[event.Type]
	idx.addToIndex(&eventList, event)
	idx.typeIndex[event.Type] = eventList

	// Extract PID and other fields from event data
	var pid uint32
	var cgroupID uint64
	var processName string

	switch event.Type {
	case events.EventTypeExec:
		switch v := event.Data.(type) {
		case *events.ExecEvent:
			pid = v.Hdr.PID
			cgroupID = v.Hdr.CgroupID
			processName = extractCString(v.Hdr.Comm[:])
		case events.ExecEvent:
			pid = v.Hdr.PID
			cgroupID = v.Hdr.CgroupID
			processName = extractCString(v.Hdr.Comm[:])
		}
	case events.EventTypeFileOpen:
		switch v := event.Data.(type) {
		case *events.FileOpenEvent:
			pid = v.Hdr.PID
			cgroupID = v.Hdr.CgroupID
		case events.FileOpenEvent:
			pid = v.Hdr.PID
			cgroupID = v.Hdr.CgroupID
		}
	case events.EventTypeConnect:
		switch v := event.Data.(type) {
		case *events.ConnectEvent:
			pid = v.Hdr.PID
			cgroupID = v.Hdr.CgroupID
		case events.ConnectEvent:
			pid = v.Hdr.PID
			cgroupID = v.Hdr.CgroupID
		}
	}

	// Index by PID
	if pid != 0 {
		eventList := idx.pidIndex[pid]
		idx.addToIndex(&eventList, event)
		idx.pidIndex[pid] = eventList
	}

	// Index by CgroupID
	if cgroupID != 0 {
		eventList := idx.cgroupIndex[cgroupID]
		idx.addToIndex(&eventList, event)
		idx.cgroupIndex[cgroupID] = eventList
	}

	// Index by process name
	if processName != "" {
		eventList := idx.processIndex[processName]
		idx.addToIndex(&eventList, event)
		idx.processIndex[processName] = eventList
	}
}

// addToIndex adds an event to an index slice, maintaining size limit.
func (idx *Indexer) addToIndex(slice *[]*Event, event *Event) {
	*slice = append(*slice, event)
	if len(*slice) > idx.maxIndexSize {
		// Remove oldest event (keep most recent)
		*slice = (*slice)[1:]
	}
}

func (idx *Indexer) QueryByPID(pid uint32) []*Event {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	events := idx.pidIndex[pid]
	result := make([]*Event, len(events))
	copy(result, events)
	return result
}

func (idx *Indexer) QueryByCgroup(cgroupID uint64) []*Event {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	events := idx.cgroupIndex[cgroupID]
	result := make([]*Event, len(events))
	copy(result, events)
	return result
}

func (idx *Indexer) QueryByType(eventType events.EventType) []*Event {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	events := idx.typeIndex[eventType]
	result := make([]*Event, len(events))
	copy(result, events)
	return result
}

func (idx *Indexer) QueryByProcess(processName string) []*Event {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	events := idx.processIndex[processName]
	result := make([]*Event, len(events))
	copy(result, events)
	return result
}

func (idx *Indexer) QueryByFilter(filter Filter) []*Event {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var candidateSets []map[*Event]bool

	// Filter by types
	if len(filter.Types) > 0 {
		set := make(map[*Event]bool)
		for _, eventType := range filter.Types {
			for _, event := range idx.typeIndex[eventType] {
				set[event] = true
			}
		}
		if len(set) > 0 {
			candidateSets = append(candidateSets, set)
		}
	}

	// Filter by PIDs
	if len(filter.PIDs) > 0 {
		set := make(map[*Event]bool)
		for _, pid := range filter.PIDs {
			for _, event := range idx.pidIndex[pid] {
				set[event] = true
			}
		}
		if len(set) > 0 {
			candidateSets = append(candidateSets, set)
		}
	}

	// Filter by CgroupIDs
	if len(filter.CgroupIDs) > 0 {
		set := make(map[*Event]bool)
		for _, cgroupID := range filter.CgroupIDs {
			for _, event := range idx.cgroupIndex[cgroupID] {
				set[event] = true
			}
		}
		if len(set) > 0 {
			candidateSets = append(candidateSets, set)
		}
	}

	// Filter by process names
	if len(filter.Processes) > 0 {
		set := make(map[*Event]bool)
		for _, processName := range filter.Processes {
			for _, event := range idx.processIndex[processName] {
				set[event] = true
			}
		}
		if len(set) > 0 {
			candidateSets = append(candidateSets, set)
		}
	}

	// If no filters specified, return empty
	if len(candidateSets) == 0 {
		return []*Event{}
	}

	// Intersect all candidate sets (AND operation)
	resultSet := candidateSets[0]
	for i := 1; i < len(candidateSets); i++ {
		for event := range resultSet {
			if !candidateSets[i][event] {
				delete(resultSet, event)
			}
		}
	}

	// Convert to slice
	result := make([]*Event, 0, len(resultSet))
	for event := range resultSet {
		result = append(result, event)
	}

	return result
}

func (idx *Indexer) Cleanup(validEvents map[*Event]bool) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Clean up PID index
	for pid, events := range idx.pidIndex {
		filtered := make([]*Event, 0, len(events))
		for _, event := range events {
			if validEvents[event] {
				filtered = append(filtered, event)
			}
		}
		if len(filtered) == 0 {
			delete(idx.pidIndex, pid)
		} else {
			idx.pidIndex[pid] = filtered
		}
	}

	// Clean up CgroupID index
	for cgroupID, events := range idx.cgroupIndex {
		filtered := make([]*Event, 0, len(events))
		for _, event := range events {
			if validEvents[event] {
				filtered = append(filtered, event)
			}
		}
		if len(filtered) == 0 {
			delete(idx.cgroupIndex, cgroupID)
		} else {
			idx.cgroupIndex[cgroupID] = filtered
		}
	}

	// Clean up type index
	for eventType, events := range idx.typeIndex {
		filtered := make([]*Event, 0, len(events))
		for _, event := range events {
			if validEvents[event] {
				filtered = append(filtered, event)
			}
		}
		if len(filtered) == 0 {
			delete(idx.typeIndex, eventType)
		} else {
			idx.typeIndex[eventType] = filtered
		}
	}

	// Clean up process index
	for processName, events := range idx.processIndex {
		filtered := make([]*Event, 0, len(events))
		for _, event := range events {
			if validEvents[event] {
				filtered = append(filtered, event)
			}
		}
		if len(filtered) == 0 {
			delete(idx.processIndex, processName)
		} else {
			idx.processIndex[processName] = filtered
		}
	}
}

// extractCString extracts a null-terminated C string from a byte array.
func extractCString(data []byte) string {
	for i, b := range data {
		if b == 0 {
			return string(data[:i])
		}
	}
	return string(data)
}
