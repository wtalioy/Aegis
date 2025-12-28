package proc

import (
	"sync"
	"sync/atomic"
	"time"
)

type ProcessProfile struct {
	PID      uint32
	Static   StaticProfile
	Dynamic  DynamicProfile
	Baseline *BaselineProfile // Optional baseline for anomaly detection
	mu       sync.RWMutex
}

type StaticProfile struct {
	StartTime   time.Time
	CommandLine string
	Genealogy   []uint32 // Parent process chain
}

type DynamicProfile struct {
	FileOpenCount   int64
	NetConnectCount int64
	ExecCount       int64
	LastFileOpen    time.Time
	LastConnect     time.Time
	LastExec        time.Time
}

type BaselineProfile struct {
	NormalFileRate     float64 // Files per minute
	NormalNetRate      float64 // Connections per minute
	NormalExecRate     float64 // Execs per minute
	CommonFilePatterns []string
	CommonNetPorts     []uint16
}

// ProfileRegistry manages process profiles.
type ProfileRegistry struct {
	profiles sync.Map // map[uint32]*ProcessProfile
}

func NewProfileRegistry() *ProfileRegistry {
	return &ProfileRegistry{}
}

func (pr *ProfileRegistry) GetProfile(pid uint32) (*ProcessProfile, bool) {
	profile, ok := pr.profiles.Load(pid)
	if !ok {
		return nil, false
	}
	return profile.(*ProcessProfile), true
}

func (pr *ProfileRegistry) GetOrCreateProfile(pid uint32, startTime time.Time, commandLine string, genealogy []uint32) *ProcessProfile {
	profile, ok := pr.profiles.Load(pid)
	if ok {
		return profile.(*ProcessProfile)
	}

	newProfile := &ProcessProfile{
		PID: pid,
		Static: StaticProfile{
			StartTime:   startTime,
			CommandLine: commandLine,
			Genealogy:   genealogy,
		},
	}

	actual, _ := pr.profiles.LoadOrStore(pid, newProfile)
	return actual.(*ProcessProfile)
}

// RecordFileOpen records a file open event for a process.
func (pr *ProfileRegistry) RecordFileOpen(pid uint32) {
	profile, ok := pr.GetProfile(pid)
	if !ok {
		return
	}

	profile.mu.Lock()
	atomic.AddInt64(&profile.Dynamic.FileOpenCount, 1)
	profile.Dynamic.LastFileOpen = time.Now()
	profile.mu.Unlock()
}

// RecordConnect records a network connection event for a process.
func (pr *ProfileRegistry) RecordConnect(pid uint32) {
	profile, ok := pr.GetProfile(pid)
	if !ok {
		return
	}

	profile.mu.Lock()
	atomic.AddInt64(&profile.Dynamic.NetConnectCount, 1)
	profile.Dynamic.LastConnect = time.Now()
	profile.mu.Unlock()
}

// RecordExec records an exec event for a process.
func (pr *ProfileRegistry) RecordExec(pid uint32) {
	profile, ok := pr.GetProfile(pid)
	if !ok {
		return
	}

	profile.mu.Lock()
	atomic.AddInt64(&profile.Dynamic.ExecCount, 1)
	profile.Dynamic.LastExec = time.Now()
	profile.mu.Unlock()
}

func (pr *ProfileRegistry) GetAnomalousProcesses() []*ProcessProfile {
	var anomalous []*ProcessProfile

	pr.profiles.Range(func(key, value any) bool {
		profile := value.(*ProcessProfile)
		if pr.isAnomalous(profile) {
			anomalous = append(anomalous, profile)
		}
		return true
	})

	return anomalous
}

// isAnomalous checks if a process shows anomalous behavior.
func (pr *ProfileRegistry) isAnomalous(profile *ProcessProfile) bool {
	profile.mu.RLock()
	defer profile.mu.RUnlock()

	if profile.Baseline == nil {
		return false // No baseline to compare against
	}

	// Simple anomaly detection: check if current rates exceed baseline significantly
	// This is a simplified version; Phase 3 will add AI-powered anomaly detection

	// Calculate current rates (simplified - should use time windows)
	fileRate := float64(profile.Dynamic.FileOpenCount)
	netRate := float64(profile.Dynamic.NetConnectCount)

	// Check if rates are significantly higher than baseline (3x threshold)
	if profile.Baseline.NormalFileRate > 0 && fileRate > profile.Baseline.NormalFileRate*3 {
		return true
	}
	if profile.Baseline.NormalNetRate > 0 && netRate > profile.Baseline.NormalNetRate*3 {
		return true
	}

	return false
}

func (pr *ProfileRegistry) RemoveProfile(pid uint32) {
	pr.profiles.Delete(pid)
}
