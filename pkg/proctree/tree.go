package proctree

import (
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ProcessInfo struct {
	PID       uint32
	PPID      uint32
	CgroupID  uint64
	Comm      string
	Timestamp time.Time
}

type ProcessTree struct {
	processes      sync.Map
	timeIndex      *timeIndex
	maxAge         time.Duration
	maxSize        int
	maxChainLength int
	size           atomic.Int32
}

func NewProcessTree(maxAge time.Duration, maxSize int, maxChainLength int) *ProcessTree {
	pt := &ProcessTree{
		timeIndex:      newTimeIndex(),
		maxAge:         maxAge,
		maxSize:        maxSize,
		maxChainLength: maxChainLength,
	}

	go func() {
		if err := pt.seedFromProc(); err != nil {
			log.Printf("Warning: failed to seed process tree from /proc: %v", err)
		}
	}()

	go pt.cleanupLoop()

	return pt
}

func (pt *ProcessTree) seedFromProc() error {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.ParseUint(entry.Name(), 10, 32)
		if err != nil {
			continue
		}

		info, err := pt.readProcInfo(uint32(pid))
		if err != nil {
			continue
		}

		pt.processes.Store(uint32(pid), info)
		pt.timeIndex.Add(uint32(pid), info.Timestamp)
		pt.size.Add(1)
		count++
	}

	log.Printf("Process tree seeded with %d processes from /proc", count)
	return nil
}

func (pt *ProcessTree) readProcInfo(pid uint32) (*ProcessInfo, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return nil, err
	}

	// pid (comm) state ppid ...
	str := string(data)

	commEnd := strings.LastIndex(str, ")")
	if commEnd == -1 {
		return nil, fmt.Errorf("invalid stat format")
	}
	commStart := strings.Index(str, "(")
	if commStart == -1 {
		return nil, fmt.Errorf("invalid stat format")
	}
	comm := str[commStart+1 : commEnd]

	fields := strings.Fields(str[commEnd+1:])
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid stat format")
	}
	ppid, err := strconv.ParseUint(fields[1], 10, 32)
	if err != nil {
		return nil, err
	}

	cgroupID := pt.readCgroupID(pid)

	return &ProcessInfo{
		PID:       pid,
		PPID:      uint32(ppid),
		CgroupID:  cgroupID,
		Comm:      comm,
		Timestamp: time.Now(),
	}, nil
}

func (pt *ProcessTree) readCgroupID(pid uint32) uint64 {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		return 1 // Default to host
	}

	lines := strings.Split(string(data), "\n")

	// cgroup v2 format: 0::/path
	for _, line := range lines {
		if after, ok := strings.CutPrefix(line, "0::"); ok {
			cgroupPath := after
			if cgroupPath == "/" || cgroupPath == "" {
				return 1 // Host process
			}
			return hashString(cgroupPath)
		}
	}

	// If no cgroup v2, try v1 (look for docker/containerd patterns)
	for _, line := range lines {
		if strings.Contains(line, "/docker/") || strings.Contains(line, "/containerd/") {
			parts := strings.SplitN(line, ":", 3)
			if len(parts) == 3 {
				return hashString(parts[2])
			}
		}
	}

	return 1
}

func hashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func (pt *ProcessTree) AddProcess(pid, ppid uint32, cgroupID uint64, comm string) {
	if pt.size.Load() >= int32(pt.maxSize) {
		pt.evictOldest()
	}

	info := &ProcessInfo{
		PID:       pid,
		PPID:      ppid,
		CgroupID:  cgroupID,
		Comm:      comm,
		Timestamp: time.Now(),
	}

	if _, exists := pt.processes.Load(pid); !exists {
		pt.size.Add(1)
	}

	pt.processes.Store(pid, info)
	pt.timeIndex.Add(pid, info.Timestamp)
}

func (pt *ProcessTree) evictOldest() {
	oldestPID, ok := pt.timeIndex.PopOldest()
	if !ok {
		return
	}

	if _, loaded := pt.processes.LoadAndDelete(oldestPID); loaded {
		pt.size.Add(-1)
	}
}

func (pt *ProcessTree) GetProcess(pid uint32) (*ProcessInfo, bool) {
	val, ok := pt.processes.Load(pid)
	if !ok {
		return nil, false
	}
	return val.(*ProcessInfo), true
}

func (pt *ProcessTree) GetAncestors(pid uint32) []*ProcessInfo {
	chain := make([]*ProcessInfo, 0, pt.maxChainLength)
	visited := make(map[uint32]bool)

	for currentPID := pid; currentPID != 0 && currentPID != 1 && len(chain) < pt.maxChainLength; {
		if visited[currentPID] {
			break
		}
		visited[currentPID] = true

		info, ok := pt.GetProcess(currentPID)
		if !ok {
			break
		}
		chain = append(chain, info)

		// Stop at container boundary (cgroup change)
		if len(chain) > 1 && info.CgroupID != chain[0].CgroupID {
			break
		}

		currentPID = info.PPID
	}

	return chain
}

func (pt *ProcessTree) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		pt.cleanup()
	}
}

func (pt *ProcessTree) cleanup() {
	now := time.Now()
	count := 0

	for {
		pid, timestamp, ok := pt.timeIndex.GetOldest()
		if !ok {
			break
		}
		if now.Sub(timestamp) <= pt.maxAge {
			break
		}
		pt.timeIndex.PopOldest()
		if _, loaded := pt.processes.LoadAndDelete(pid); loaded {
			pt.size.Add(-1)
			count++
		}
	}

	if count > 0 {
		log.Printf("Cleaned up %d old processes from tree (current size: %d)", count, pt.size.Load())
	}
}
