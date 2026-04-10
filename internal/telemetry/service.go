package telemetry

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"aegis/internal/platform/events"
	"aegis/internal/platform/storage"
	"aegis/internal/shared/utils"
	"aegis/internal/telemetry/proc"
	"aegis/internal/telemetry/workload"
)

type EventType string

const (
	EventTypeExec    EventType = "exec"
	EventTypeFile    EventType = "file"
	EventTypeConnect EventType = "connect"
)

type Event struct {
	ID          string    `json:"id"`
	Type        EventType `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	PID         uint32    `json:"pid"`
	PPID        uint32    `json:"ppid,omitempty"`
	CgroupID    uint64    `json:"cgroup_id"`
	ProcessName string    `json:"process_name"`
	ParentName  string    `json:"parent_name,omitempty"`
	CommandLine string    `json:"command_line,omitempty"`
	Filename    string    `json:"filename,omitempty"`
	Flags       uint32    `json:"flags,omitempty"`
	Ino         uint64    `json:"ino,omitempty"`
	Dev         uint64    `json:"dev,omitempty"`
	Family      uint16    `json:"family,omitempty"`
	Port        uint16    `json:"port,omitempty"`
	Address     string    `json:"address,omitempty"`
	Blocked     bool      `json:"blocked"`
}

type Record struct {
	Event Event
	Raw   *storage.Event
}

type Filter struct {
	Types     []EventType
	Processes []string
	PIDs      []uint32
	CgroupIDs []uint64
	Start     *time.Time
	End       *time.Time
}

type Query struct {
	Filter Filter
	Page   int
	Limit  int
}

type TypeCounts struct {
	Exec    int `json:"exec"`
	File    int `json:"file"`
	Connect int `json:"connect"`
}

type PageResult struct {
	Events     []Event    `json:"events"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	TotalPages int        `json:"total_pages"`
	TypeCounts TypeCounts `json:"type_counts"`
}

type Service struct {
	mu          sync.RWMutex
	records     []*Record
	recordsByID map[string]*Record
	capacity    int

	rawStore    *storage.Manager
	processTree *proc.ProcessTree
	workloads   *workload.Registry
	profiles    *proc.ProfileRegistry
}

func NewService(capacity int, indexSize int, processTree *proc.ProcessTree, workloads *workload.Registry, profiles *proc.ProfileRegistry) *Service {
	if capacity <= 0 {
		capacity = 10000
	}
	if indexSize <= 0 {
		indexSize = 1000
	}
	return &Service{
		records:     make([]*Record, 0, capacity),
		recordsByID: make(map[string]*Record),
		capacity:    capacity,
		rawStore:    storage.NewManager(capacity, indexSize),
		processTree: processTree,
		workloads:   workloads,
		profiles:    profiles,
	}
}

func (s *Service) Ingest(record *events.DecodedRecord) (*Record, error) {
	if record == nil {
		return nil, fmt.Errorf("decoded record is nil")
	}

	switch {
	case record.Exec != nil:
		return s.ingestExec(record)
	case record.FileOpen != nil:
		return s.ingestFile(record)
	case record.Connect != nil:
		return s.ingestConnect(record)
	default:
		return nil, fmt.Errorf("decoded record has no event payload")
	}
}

func (s *Service) Latest(limit int) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.records) {
		limit = len(s.records)
	}

	result := make([]Event, 0, limit)
	for i := len(s.records) - 1; i >= 0 && len(result) < limit; i-- {
		result = append(result, s.records[i].Event)
	}
	return result
}

func (s *Service) Get(id string) (*Record, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.recordsByID[id]
	if !ok {
		return nil, false
	}
	copyRecord := *record
	return &copyRecord, true
}

func (s *Service) Query(req Query) PageResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]Event, 0, len(s.records))
	counts := TypeCounts{}
	for _, record := range s.records {
		if !matches(record.Event, req.Filter) {
			continue
		}
		filtered = append(filtered, record.Event)
		switch record.Event.Type {
		case EventTypeExec:
			counts.Exec++
		case EventTypeFile:
			counts.File++
		case EventTypeConnect:
			counts.Connect++
		}
	}

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

	total := len(filtered)
	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return PageResult{
		Events:     append([]Event(nil), filtered[start:end]...),
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		TypeCounts: counts,
	}
}

func (s *Service) RawStore() storage.EventStore {
	return s.rawStore
}

func (s *Service) ProcessTree() *proc.ProcessTree {
	return s.processTree
}

func (s *Service) Workloads() *workload.Registry {
	return s.workloads
}

func (s *Service) Profiles() *proc.ProfileRegistry {
	return s.profiles
}

func (s *Service) RecordAlert(cgroupID uint64, blocked bool) {
	if s.workloads != nil && cgroupID != 0 {
		s.workloads.RecordAlert(cgroupID, blocked)
	}
}

func (s *Service) ingestExec(record *events.DecodedRecord) (*Record, error) {
	ev := *record.Exec
	processName := utils.ExtractCString(ev.Hdr.Comm[:])
	parentName := utils.ExtractCString(ev.PComm[:])
	commandLine := utils.ExtractCString(ev.CommandLine[:])
	if s.processTree != nil {
		s.processTree.AddProcess(ev.Hdr.PID, ev.PPID, ev.Hdr.CgroupID, processName)
	}
	if s.workloads != nil {
		s.workloads.RecordExec(ev.Hdr.CgroupID, proc.ResolveCgroupPath(ev.Hdr.PID, ev.Hdr.CgroupID))
	}
	if s.profiles != nil {
		var genealogy []uint32
		if s.processTree != nil {
			ancestors := s.processTree.GetAncestors(ev.Hdr.PID)
			genealogy = make([]uint32, 0, len(ancestors))
			for _, ancestor := range ancestors {
				if ancestor == nil {
					continue
				}
				genealogy = append(genealogy, ancestor.PID)
			}
		}
		s.profiles.GetOrCreateProfile(ev.Hdr.PID, ev.Hdr.Timestamp(), commandLine, genealogy)
		s.profiles.RecordExec(ev.Hdr.PID)
	}

	raw := storage.EventFromBackend(events.EventTypeExec, ev.Hdr.Timestamp(), ev)
	_ = s.rawStore.Append(raw)

	event := Event{
		Type:        EventTypeExec,
		Timestamp:   ev.Hdr.Timestamp(),
		PID:         ev.Hdr.PID,
		PPID:        ev.PPID,
		CgroupID:    ev.Hdr.CgroupID,
		ProcessName: processName,
		ParentName:  parentName,
		CommandLine: commandLine,
		Blocked:     ev.Hdr.Blocked == 1,
	}
	event.ID = generateEventID(raw)

	return s.appendRecord(event, raw), nil
}

func (s *Service) ingestFile(record *events.DecodedRecord) (*Record, error) {
	ev := *record.FileOpen
	processName := utils.ExtractCString(ev.Hdr.Comm[:])
	if s.processTree != nil {
		if info, ok := s.processTree.GetProcess(ev.Hdr.PID); ok && info.Comm != "" {
			processName = info.Comm
		}
	}
	if s.workloads != nil {
		s.workloads.RecordFile(ev.Hdr.CgroupID, proc.ResolveCgroupPath(ev.Hdr.PID, ev.Hdr.CgroupID))
	}
	if s.profiles != nil {
		s.profiles.RecordFileOpen(ev.Hdr.PID)
	}

	raw := storage.EventFromBackend(events.EventTypeFileOpen, ev.Hdr.Timestamp(), ev)
	_ = s.rawStore.Append(raw)

	event := Event{
		Type:        EventTypeFile,
		Timestamp:   ev.Hdr.Timestamp(),
		PID:         ev.Hdr.PID,
		CgroupID:    ev.Hdr.CgroupID,
		ProcessName: processName,
		Filename:    utils.ExtractCString(ev.Filename[:]),
		Flags:       ev.Flags,
		Ino:         ev.Ino,
		Dev:         ev.Dev,
		Blocked:     ev.Hdr.Blocked == 1,
	}
	event.ID = generateEventID(raw)

	return s.appendRecord(event, raw), nil
}

func (s *Service) ingestConnect(record *events.DecodedRecord) (*Record, error) {
	ev := *record.Connect
	processName := utils.ExtractCString(ev.Hdr.Comm[:])
	if s.processTree != nil {
		if info, ok := s.processTree.GetProcess(ev.Hdr.PID); ok && info.Comm != "" {
			processName = info.Comm
		}
	}
	if s.workloads != nil {
		s.workloads.RecordConnect(ev.Hdr.CgroupID, proc.ResolveCgroupPath(ev.Hdr.PID, ev.Hdr.CgroupID))
	}
	if s.profiles != nil {
		s.profiles.RecordConnect(ev.Hdr.PID)
	}

	raw := storage.EventFromBackend(events.EventTypeConnect, ev.Hdr.Timestamp(), ev)
	_ = s.rawStore.Append(raw)

	ip := utils.ExtractIP(&ev)
	address := fmt.Sprintf("%s:%d", ip, ev.Port)
	event := Event{
		Type:        EventTypeConnect,
		Timestamp:   ev.Hdr.Timestamp(),
		PID:         ev.Hdr.PID,
		CgroupID:    ev.Hdr.CgroupID,
		ProcessName: processName,
		Family:      ev.Family,
		Port:        ev.Port,
		Address:     address,
		Blocked:     ev.Hdr.Blocked == 1,
	}
	event.ID = generateEventID(raw)

	return s.appendRecord(event, raw), nil
}

func (s *Service) appendRecord(event Event, raw *storage.Event) *Record {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := &Record{Event: event, Raw: raw}
	s.records = append(s.records, record)
	s.recordsByID[event.ID] = record
	if len(s.records) > s.capacity {
		oldest := s.records[0]
		delete(s.recordsByID, oldest.Event.ID)
		s.records = s.records[1:]
	}
	return record
}

func generateEventID(event *storage.Event) string {
	h := sha256.New()
	h.Write([]byte(event.Timestamp.Format(time.RFC3339Nano)))
	fmt.Fprintf(h, "%d", int(event.Type))

	switch ev := event.Data.(type) {
	case *events.ExecEvent:
		h.Write(ev.Hdr.Comm[:])
		fmt.Fprintf(h, "%d", ev.Hdr.PID)
	case events.ExecEvent:
		h.Write(ev.Hdr.Comm[:])
		fmt.Fprintf(h, "%d", ev.Hdr.PID)
	case *events.FileOpenEvent:
		h.Write(ev.Filename[:])
		fmt.Fprintf(h, "%d", ev.Hdr.PID)
	case events.FileOpenEvent:
		h.Write(ev.Filename[:])
		fmt.Fprintf(h, "%d", ev.Hdr.PID)
	case *events.ConnectEvent:
		fmt.Fprintf(h, "%d:%d", ev.Port, ev.Hdr.PID)
	case events.ConnectEvent:
		fmt.Fprintf(h, "%d:%d", ev.Port, ev.Hdr.PID)
	}

	return hex.EncodeToString(h.Sum(nil))[:16]
}

func matches(event Event, filter Filter) bool {
	if len(filter.Types) > 0 {
		typeMatched := false
		for _, eventType := range filter.Types {
			if event.Type == eventType {
				typeMatched = true
				break
			}
		}
		if !typeMatched {
			return false
		}
	}

	if filter.Start != nil && event.Timestamp.Before(*filter.Start) {
		return false
	}
	if filter.End != nil && event.Timestamp.After(*filter.End) {
		return false
	}

	if len(filter.PIDs) > 0 {
		pidMatched := false
		for _, pid := range filter.PIDs {
			if event.PID == pid {
				pidMatched = true
				break
			}
		}
		if !pidMatched {
			return false
		}
	}

	if len(filter.CgroupIDs) > 0 {
		cgroupMatched := false
		for _, cgroupID := range filter.CgroupIDs {
			if event.CgroupID == cgroupID {
				cgroupMatched = true
				break
			}
		}
		if !cgroupMatched {
			return false
		}
	}

	if len(filter.Processes) > 0 {
		processMatched := false
		for _, processName := range filter.Processes {
			if strings.Contains(event.ProcessName, processName) || strings.Contains(processName, event.ProcessName) {
				processMatched = true
				break
			}
		}
		if !processMatched {
			return false
		}
	}

	return true
}
