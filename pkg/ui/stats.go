package ui

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"eulerguard/pkg/events"
	"eulerguard/pkg/types"
	"eulerguard/pkg/utils"
)

type WorkloadCountFunc func() int

type Stats struct {
	execCount    atomic.Int64
	fileCount    atomic.Int64
	connectCount atomic.Int64

	lastSecExec    atomic.Int64
	lastSecFile    atomic.Int64
	lastSecConnect atomic.Int64
	rateExec       atomic.Int64
	rateFile       atomic.Int64
	rateConnect    atomic.Int64

	alerts      []types.Alert
	alertsMu    sync.RWMutex
	maxAlerts   int
	totalAlerts atomic.Int64
	alertDedup  map[alertKey]time.Time
	dedupWindow time.Duration

	recentExecs    []types.ExecEvent
	recentFiles    []types.FileEvent
	recentConnects []types.ConnectEvent
	recentMu       sync.RWMutex
	maxRecent      int

	workloadCountFn WorkloadCountFunc

	eventSubs   map[chan any]struct{}
	eventSubsMu sync.RWMutex
}

type alertKey struct {
	RuleName    string
	ProcessName string
	CgroupID    string
	Action      string
}

type sseEvent struct {
	Name string
	Data any
}

func NewStats() *Stats {
	s := &Stats{
		alerts:         make([]types.Alert, 0, 100),
		maxAlerts:      100,
		recentExecs:    make([]types.ExecEvent, 0, 50),
		recentFiles:    make([]types.FileEvent, 0, 50),
		recentConnects: make([]types.ConnectEvent, 0, 50),
		maxRecent:      50,
		eventSubs:      make(map[chan any]struct{}),
		alertDedup:     make(map[alertKey]time.Time),
		dedupWindow:    10 * time.Second,
	}
	go s.rateLoop()
	return s
}

func (s *Stats) SetWorkloadCountFunc(fn WorkloadCountFunc) {
	s.workloadCountFn = fn
}

func (s *Stats) rateLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		exec := s.lastSecExec.Swap(0)
		file := s.lastSecFile.Swap(0)
		net := s.lastSecConnect.Swap(0)

		s.rateExec.Store(exec)
		s.rateFile.Store(file)
		s.rateConnect.Store(net)
	}
}

func (s *Stats) RecordExec(ev events.ExecEvent) {
	s.execCount.Add(1)
	s.lastSecExec.Add(1)

	frontendEv := ExecToFrontend(ev)
	s.recentMu.Lock()
	if len(s.recentExecs) >= s.maxRecent {
		s.recentExecs = s.recentExecs[1:]
	}
	s.recentExecs = append(s.recentExecs, frontendEv)
	s.recentMu.Unlock()
}

func (s *Stats) RecordFile() { s.fileCount.Add(1); s.lastSecFile.Add(1) }

func (s *Stats) RecordFileEvent(ev types.FileEvent) {
	s.fileCount.Add(1)
	s.lastSecFile.Add(1)

	s.recentMu.Lock()
	if len(s.recentFiles) >= s.maxRecent {
		s.recentFiles = s.recentFiles[1:]
	}
	s.recentFiles = append(s.recentFiles, ev)
	s.recentMu.Unlock()
}

func (s *Stats) RecordConnect() { s.connectCount.Add(1); s.lastSecConnect.Add(1) }

func (s *Stats) RecordConnectEvent(ev types.ConnectEvent) {
	s.connectCount.Add(1)
	s.lastSecConnect.Add(1)

	s.recentMu.Lock()
	if len(s.recentConnects) >= s.maxRecent {
		s.recentConnects = s.recentConnects[1:]
	}
	s.recentConnects = append(s.recentConnects, ev)
	s.recentMu.Unlock()
}

func (s *Stats) AddAlert(alert types.Alert) {
	s.alertsMu.Lock()
	now := time.Now()
	if s.dedupWindow > 0 {
		s.purgeDedupLocked(now)
		key := alertKey{
			RuleName:    alert.RuleName,
			ProcessName: alert.ProcessName,
			CgroupID:    alert.CgroupID,
			Action:      alert.Action,
		}
		if last, ok := s.alertDedup[key]; ok && now.Sub(last) < s.dedupWindow {
			s.alertsMu.Unlock()
			return
		}
		s.alertDedup[key] = now
	}
	if len(s.alerts) >= s.maxAlerts {
		s.alerts = s.alerts[1:]
	}
	s.alerts = append(s.alerts, alert)
	s.alertsMu.Unlock()
	s.totalAlerts.Add(1)
}

func (s *Stats) purgeDedupLocked(now time.Time) {
	if len(s.alertDedup) == 0 || s.dedupWindow <= 0 {
		return
	}
	expireBefore := now.Add(-s.dedupWindow)
	for key, ts := range s.alertDedup {
		if ts.Before(expireBefore) {
			delete(s.alertDedup, key)
		}
	}
}

func (s *Stats) Rates() (exec, file, net int64) {
	return s.rateExec.Load(), s.rateFile.Load(), s.rateConnect.Load()
}

func (s *Stats) Counts() (exec, file, net int64) {
	return s.execCount.Load(), s.fileCount.Load(), s.connectCount.Load()
}

func (s *Stats) AlertCount() int {
	s.alertsMu.RLock()
	defer s.alertsMu.RUnlock()
	return len(s.alerts)
}

func (s *Stats) TotalAlertCount() int64 {
	return s.totalAlerts.Load()
}

func (s *Stats) Alerts() []types.Alert {
	s.alertsMu.RLock()
	defer s.alertsMu.RUnlock()
	result := make([]types.Alert, len(s.alerts))
	copy(result, s.alerts)
	return result
}

func (s *Stats) WorkloadCount() int {
	if s.workloadCountFn != nil {
		return s.workloadCountFn()
	}
	return 0
}

func (s *Stats) RecentExecs() []types.ExecEvent {
	s.recentMu.RLock()
	defer s.recentMu.RUnlock()
	result := make([]types.ExecEvent, len(s.recentExecs))
	copy(result, s.recentExecs)
	return result
}

func (s *Stats) RecentFiles() []types.FileEvent {
	s.recentMu.RLock()
	defer s.recentMu.RUnlock()
	result := make([]types.FileEvent, len(s.recentFiles))
	copy(result, s.recentFiles)
	return result
}

func (s *Stats) RecentConnects() []types.ConnectEvent {
	s.recentMu.RLock()
	defer s.recentMu.RUnlock()
	result := make([]types.ConnectEvent, len(s.recentConnects))
	copy(result, s.recentConnects)
	return result
}

var _ types.StatsProvider = (*Stats)(nil)

func ExecToFrontend(ev events.ExecEvent) types.ExecEvent {
	return types.ExecEvent{
		Type:       "exec",
		Timestamp:  time.Now().UnixMilli(),
		PID:        ev.PID,
		PPID:       ev.PPID,
		CgroupID:   strconv.FormatUint(ev.CgroupID, 10),
		Comm:       utils.ExtractCString(ev.Comm[:]),
		ParentComm: utils.ExtractCString(ev.PComm[:]),
		Blocked:    ev.Blocked == 1,
	}
}

func FileToFrontend(ev events.FileOpenEvent, filename string) types.FileEvent {
	return types.FileEvent{
		Type:      "file",
		Timestamp: time.Now().UnixMilli(),
		PID:       ev.PID,
		CgroupID:  strconv.FormatUint(ev.CgroupID, 10),
		Flags:     ev.Flags,
		Ino:       ev.Ino,
		Dev:       ev.Dev,
		Filename:  filename,
		Blocked:   ev.Blocked == 1,
	}
}

func ConnectToFrontend(ev events.ConnectEvent, addr string, processName string) types.ConnectEvent {
	return types.ConnectEvent{
		Type:        "connect",
		Timestamp:   time.Now().UnixMilli(),
		PID:         ev.PID,
		ProcessName: processName,
		CgroupID:    strconv.FormatUint(ev.CgroupID, 10),
		Family:      ev.Family,
		Port:        ev.Port,
		Addr:        addr,
		Blocked:     ev.Blocked == 1,
	}
}

func (s *Stats) SubscribeEvents(ch chan any) {
	s.eventSubsMu.Lock()
	s.eventSubs[ch] = struct{}{}
	s.eventSubsMu.Unlock()
}

func (s *Stats) UnsubscribeEvents(ch chan any) {
	s.eventSubsMu.Lock()
	delete(s.eventSubs, ch)
	s.eventSubsMu.Unlock()
}

func (s *Stats) PublishEvent(event any) {
	s.eventSubsMu.RLock()
	defer s.eventSubsMu.RUnlock()

	for ch := range s.eventSubs {
		select {
		case ch <- event:
		default:
		}
	}
}

func (s *Stats) PublishNamedEvent(name string, data any) {
	s.PublishEvent(sseEvent{Name: name, Data: data})
}
