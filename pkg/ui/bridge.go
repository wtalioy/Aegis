package ui

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"eulerguard/pkg/events"
	"eulerguard/pkg/proc"
	"eulerguard/pkg/profiler"
	"eulerguard/pkg/rules"
	"eulerguard/pkg/utils"
	"eulerguard/pkg/workload"
)

type Bridge struct {
	stats            *Stats
	processTree      *proc.ProcessTree
	ruleEngine       *rules.Engine
	workloadRegistry *workload.Registry
	profiler         *profiler.Profiler
	mu               sync.RWMutex
}

var _ events.EventHandler = (*Bridge)(nil)

func NewBridge(stats *Stats) *Bridge {
	b := &Bridge{
		stats: stats,
	}
	return b
}

func (b *Bridge) SetRuleEngine(pt *proc.ProcessTree, re *rules.Engine) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.processTree = pt
	b.ruleEngine = re
}

func (b *Bridge) SetWorkloadRegistry(wr *workload.Registry) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.workloadRegistry = wr
}

func (b *Bridge) SetProfiler(p *profiler.Profiler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.profiler = p
}

func (b *Bridge) HandleExec(ev events.ExecEvent) {
	b.stats.RecordExec(ev)
	frontendEvent := ExecToFrontend(ev)
	b.stats.PublishEvent(frontendEvent)

	b.mu.RLock()
	re := b.ruleEngine
	prof := b.profiler
	b.mu.RUnlock()

	// Forward to profiler if active
	if prof != nil && prof.IsActive() {
		prof.HandleExec(ev)
	}

	if re == nil {
		return
	}

	comm := utils.ExtractCString(ev.Comm[:])
	pcomm := utils.ExtractCString(ev.PComm[:])

	processed := events.ProcessedEvent{
		Event:     ev,
		Timestamp: time.Now().UTC(),
		Process:   comm,
		Parent:    pcomm,
	}

	if _, _, allowed := re.MatchExec(processed); allowed {
		return
	}

	for _, alert := range re.CollectExecAlerts(processed) {
		blocked := ev.Blocked == 1
		severity := alert.Rule.Severity
		if blocked && severity != "critical" {
			severity = "critical"
		}
		b.emitAlert(FrontendAlert{
			ID:          fmt.Sprintf("exec-%d-%d", ev.PID, time.Now().UnixNano()),
			Timestamp:   time.Now().UnixMilli(),
			Severity:    severity,
			RuleName:    alert.Rule.Name,
			Description: alert.Rule.Description,
			PID:         ev.PID,
			ProcessName: comm,
			ParentName:  pcomm,
			CgroupID:    strconv.FormatUint(ev.CgroupID, 10),
			Action:      string(alert.Rule.Action),
			Blocked:     blocked,
		})
	}
}

func (b *Bridge) HandleFileOpen(ev events.FileOpenEvent, filename string) {
	frontendEvent := FileToFrontend(ev, filename)
	b.stats.RecordFileEvent(frontendEvent)
	b.stats.PublishEvent(frontendEvent)

	b.mu.RLock()
	re, pt := b.ruleEngine, b.processTree
	prof := b.profiler
	b.mu.RUnlock()

	if prof != nil && prof.IsActive() {
		prof.HandleFileOpen(ev, filename)
	}

	if re == nil {
		return
	}

	matched, rule, allowed := re.MatchFile(ev.Ino, ev.Dev, filename, ev.PID, ev.CgroupID)
	if !matched || rule == nil || allowed {
		return
	}

	var processName string
	if pt != nil {
		if info, ok := pt.GetProcess(ev.PID); ok {
			processName = info.Comm
		}
	}

	blocked := ev.Blocked == 1
	severity := rule.Severity
	if blocked && severity != "critical" {
		severity = "critical"
	}
	b.emitAlert(FrontendAlert{
		ID:          fmt.Sprintf("file-%d-%d", ev.PID, time.Now().UnixNano()),
		Timestamp:   time.Now().UnixMilli(),
		Severity:    severity,
		RuleName:    rule.Name,
		Description: fmt.Sprintf("%s: %s", rule.Description, filename),
		PID:         ev.PID,
		ProcessName: processName,
		CgroupID:    strconv.FormatUint(ev.CgroupID, 10),
		Action:      string(rule.Action),
		Blocked:     blocked,
	})
}

func (b *Bridge) HandleConnect(ev events.ConnectEvent) {
	b.mu.RLock()
	re, pt := b.ruleEngine, b.processTree
	prof := b.profiler
	b.mu.RUnlock()

	var processName string
	if pt != nil {
		if info, ok := pt.GetProcess(ev.PID); ok {
			processName = info.Comm
		}
	}

	frontendEvent := ConnectToFrontend(ev, formatAddr(ev), processName)
	b.stats.RecordConnectEvent(frontendEvent)
	b.stats.PublishEvent(frontendEvent)

	if prof != nil && prof.IsActive() {
		prof.HandleConnect(ev)
	}

	if re == nil {
		return
	}

	matched, rule, allowed := re.MatchConnect(&ev)
	if !matched || rule == nil || allowed {
		return
	}

	blocked := ev.Blocked == 1
	severity := rule.Severity
	if blocked && severity != "critical" {
		severity = "critical"
	}
	b.emitAlert(FrontendAlert{
		ID:          fmt.Sprintf("net-%d-%d", ev.PID, time.Now().UnixNano()),
		Timestamp:   time.Now().UnixMilli(),
		Severity:    severity,
		RuleName:    rule.Name,
		Description: rule.Description,
		PID:         ev.PID,
		ProcessName: processName,
		CgroupID:    strconv.FormatUint(ev.CgroupID, 10),
		Action:      string(rule.Action),
		Blocked:     blocked,
	})
}

func (b *Bridge) emitAlert(alert FrontendAlert) {
	b.stats.AddAlert(alert)
	if b.workloadRegistry != nil {
		if cgroupID, err := strconv.ParseUint(alert.CgroupID, 10, 64); err == nil {
			b.workloadRegistry.RecordAlert(cgroupID, alert.Blocked)
		}
	}
}

func (b *Bridge) NotifyRulesReload() {
	b.stats.PublishNamedEvent("rules:reload", map[string]int64{
		"timestamp": time.Now().UnixMilli(),
	})
}

func formatAddr(ev events.ConnectEvent) string {
	switch ev.Family {
	case 2:
		ip := make(net.IP, 4)
		binary.LittleEndian.PutUint32(ip, ev.AddrV4)
		return fmt.Sprintf("%s:%d", ip.String(), ev.Port)
	case 10:
		return fmt.Sprintf("[%s]:%d", net.IP(ev.AddrV6[:]).String(), ev.Port)
	default:
		return fmt.Sprintf("unknown:%d", ev.Port)
	}
}
