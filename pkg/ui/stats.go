package ui

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"eulerguard/pkg/events"
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

	alerts      []FrontendAlert
	alertsMu    sync.RWMutex
	maxAlerts   int
	totalAlerts atomic.Int64

	workloadCountFn WorkloadCountFunc

	eventSubs   map[chan any]struct{}
	eventSubsMu sync.RWMutex
}

type sseEvent struct {
	Name string
	Data any
}

func NewStats() *Stats {
	s := &Stats{
		alerts:    make([]FrontendAlert, 0, 100),
		maxAlerts: 100,
		eventSubs: make(map[chan any]struct{}),
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
}

func (s *Stats) RecordFile()    { s.fileCount.Add(1); s.lastSecFile.Add(1) }
func (s *Stats) RecordConnect() { s.connectCount.Add(1); s.lastSecConnect.Add(1) }

func (s *Stats) AddAlert(alert FrontendAlert) {
	s.totalAlerts.Add(1)
	s.alertsMu.Lock()
	if len(s.alerts) >= s.maxAlerts {
		s.alerts = s.alerts[1:]
	}
	s.alerts = append(s.alerts, alert)
	s.alertsMu.Unlock()
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

func (s *Stats) Alerts() []FrontendAlert {
	s.alertsMu.RLock()
	defer s.alertsMu.RUnlock()
	result := make([]FrontendAlert, len(s.alerts))
	copy(result, s.alerts)
	return result
}

func (s *Stats) WorkloadCount() int {
	if s.workloadCountFn != nil {
		return s.workloadCountFn()
	}
	return 0
}

func ExecToFrontend(ev events.ExecEvent) FrontendExecEvent {
	return FrontendExecEvent{
		Type:       "exec",
		Timestamp:  time.Now().UnixMilli(),
		PID:        ev.PID,
		PPID:       ev.PPID,
		CgroupID:   strconv.FormatUint(ev.CgroupID, 10),
		Comm:       utils.ExtractCString(ev.Comm[:]),
		ParentComm: utils.ExtractCString(ev.PComm[:]),
	}
}

func FileToFrontend(ev events.FileOpenEvent, filename string) FrontendFileEvent {
	return FrontendFileEvent{
		Type:      "file",
		Timestamp: time.Now().UnixMilli(),
		PID:       ev.PID,
		CgroupID:  strconv.FormatUint(ev.CgroupID, 10),
		Flags:     ev.Flags,
		Filename:  filename,
	}
}

func ConnectToFrontend(ev events.ConnectEvent, addr string) FrontendConnectEvent {
	return FrontendConnectEvent{
		Type:      "connect",
		Timestamp: time.Now().UnixMilli(),
		PID:       ev.PID,
		CgroupID:  strconv.FormatUint(ev.CgroupID, 10),
		Family:    ev.Family,
		Port:      ev.Port,
		Addr:      addr,
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

// PublishEvent sends an event to all subscribers
func (s *Stats) PublishEvent(event any) {
	s.eventSubsMu.RLock()
	defer s.eventSubsMu.RUnlock()

	for ch := range s.eventSubs {
		select {
		case ch <- event:
		default:
			// Channel full, skip this subscriber
		}
	}
}

func (s *Stats) PublishNamedEvent(name string, data any) {
	s.PublishEvent(sseEvent{Name: name, Data: data})
}
