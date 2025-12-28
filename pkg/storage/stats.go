package storage

import (
	"sync"
	"sync/atomic"
	"time"

	"aegis/pkg/apimodel"
)

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

	alerts      []apimodel.Alert
	alertsMu    sync.RWMutex
	maxAlerts   int
	totalAlerts atomic.Int64
	alertDedup  map[alertKey]time.Time
	dedupWindow time.Duration

	workloadCountFn func() int
}

type alertKey struct {
	RuleName    string
	ProcessName string
	CgroupID    string
	Action      string
}

func NewStats(maxAlerts int, dedupWindow time.Duration) *Stats {
	s := &Stats{
		alerts:      make([]apimodel.Alert, 0, maxAlerts),
		maxAlerts:   maxAlerts,
		alertDedup:  make(map[alertKey]time.Time),
		dedupWindow: dedupWindow,
	}
	go s.rateLoop()
	return s
}

func (s *Stats) SetWorkloadCountFunc(fn func() int) {
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

// RecordExec records an exec event (count only, no storage).
func (s *Stats) RecordExec() {
	s.execCount.Add(1)
	s.lastSecExec.Add(1)
}

// RecordFile records a file event (count only, no storage).
func (s *Stats) RecordFile() {
	s.fileCount.Add(1)
	s.lastSecFile.Add(1)
}

// RecordConnect records a connect event (count only, no storage).
func (s *Stats) RecordConnect() {
	s.connectCount.Add(1)
	s.lastSecConnect.Add(1)
}

func (s *Stats) AddAlert(alert apimodel.Alert) {
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

func (s *Stats) Alerts() []apimodel.Alert {
	s.alertsMu.RLock()
	defer s.alertsMu.RUnlock()
	result := make([]apimodel.Alert, len(s.alerts))
	copy(result, s.alerts)
	return result
}

func (s *Stats) WorkloadCount() int {
	if s.workloadCountFn != nil {
		return s.workloadCountFn()
	}
	return 0
}
