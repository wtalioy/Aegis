package server

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"aegis/pkg/apimodel"
	"aegis/pkg/events"
	"aegis/pkg/proc"
	"aegis/pkg/rules"
	"aegis/pkg/utils"
	"aegis/pkg/workload"
)

type Bridge struct {
	stats            *Stats
	processTree      *proc.ProcessTree
	ruleEngine       *rules.Engine
	workloadRegistry *workload.Registry
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

func (b *Bridge) HandleExec(ev events.ExecEvent) {
	b.stats.RecordExec()
	frontendEvent := ExecToFrontend(ev)
	b.stats.PublishEvent(frontendEvent)

	b.mu.RLock()
	re := b.ruleEngine
	b.mu.RUnlock()

	if re == nil {
		return
	}

	comm := utils.ExtractCString(ev.Hdr.Comm[:])
	pcomm := utils.ExtractCString(ev.PComm[:])

	processed := events.ProcessedEvent{
		Event:     ev,
		Timestamp: ev.Hdr.Timestamp(),
		Process:   comm,
		Parent:    pcomm,
	}

	blocked := ev.Hdr.Blocked == 1

	if _, _, allowed := re.MatchExec(processed); allowed {
		return
	}

	alerts := re.CollectExecAlerts(processed)

	// If kernel blocked but no alerts collected, still emit alert
	if blocked && len(alerts) == 0 {
		b.emitAlert(apimodel.Alert{
			ID:          fmt.Sprintf("exec-%d-%d", ev.Hdr.PID, time.Now().UnixNano()),
			Timestamp:   ev.Hdr.Timestamp().UnixMilli(),
			Severity:    "critical",
			RuleName:    "Kernel Blocked Execution",
			Description: fmt.Sprintf("Process execution blocked by kernel: %s", comm),
			PID:         ev.Hdr.PID,
			ProcessName: comm,
			ParentName:  pcomm,
			CgroupID:    strconv.FormatUint(ev.Hdr.CgroupID, 10),
			Action:      "block",
			Blocked:     true,
		})
		return
	}

	for _, alert := range alerts {
		// Handle testing mode: record but don't alert
		if alert.Rule.IsTesting() {
			testingBuffer := re.GetTestingBuffer()
			if testingBuffer != nil {
				testingBuffer.RecordHit(&rules.TestingHit{
					RuleName:    alert.Rule.Name,
					HitTime:     ev.Hdr.Timestamp(),
					EventType:   events.EventTypeExec,
					EventData:   &ev,
					PID:         ev.Hdr.PID,
					ProcessName: comm,
				})
			}
			continue // Skip alert for testing mode
		}

		severity := alert.Rule.Severity
		if blocked && severity != "critical" {
			severity = "critical"
		}
		b.emitAlert(apimodel.Alert{
			ID:          fmt.Sprintf("exec-%d-%d", ev.Hdr.PID, time.Now().UnixNano()),
			Timestamp:   ev.Hdr.Timestamp().UnixMilli(),
			Severity:    severity,
			RuleName:    alert.Rule.Name,
			Description: alert.Rule.Description,
			PID:         ev.Hdr.PID,
			ProcessName: comm,
			ParentName:  pcomm,
			CgroupID:    strconv.FormatUint(ev.Hdr.CgroupID, 10),
			Action:      string(alert.Rule.Action),
			Blocked:     blocked,
		})
	}
}

func (b *Bridge) HandleFileOpen(ev events.FileOpenEvent, filename string) {
	b.stats.RecordFile()
	frontendEvent := FileToFrontend(ev, filename)
	b.stats.PublishEvent(frontendEvent)

	b.mu.RLock()
	re, pt := b.ruleEngine, b.processTree
	b.mu.RUnlock()

	if re == nil {
		return
	}

	var processName string
	if pt != nil {
		if info, ok := pt.GetProcess(ev.Hdr.PID); ok {
			processName = info.Comm
		}
	}

	blocked := ev.Hdr.Blocked == 1

	matched, rule, allowed := re.MatchFile(ev.Ino, ev.Dev, filename, ev.Hdr.PID, ev.Hdr.CgroupID)

	// If kernel blocked the file but Go-side matching failed, still emit alert
	if blocked && (!matched || rule == nil) {
		b.emitAlert(apimodel.Alert{
			ID:          fmt.Sprintf("file-%d-%d", ev.Hdr.PID, time.Now().UnixNano()),
			Timestamp:   ev.Hdr.Timestamp().UnixMilli(),
			Severity:    "critical",
			RuleName:    "Kernel Blocked File Access",
			Description: fmt.Sprintf("File access blocked by kernel: %s", filename),
			PID:         ev.Hdr.PID,
			ProcessName: processName,
			CgroupID:    strconv.FormatUint(ev.Hdr.CgroupID, 10),
			Action:      "block",
			Blocked:     true,
		})
		return
	}

	if !matched || rule == nil || allowed {
		return
	}

	// Handle testing mode: record but don't alert
	if rule.IsTesting() {
		testingBuffer := re.GetTestingBuffer()
		if testingBuffer != nil {
			testingBuffer.RecordHit(&rules.TestingHit{
				RuleName:    rule.Name,
				HitTime:     ev.Hdr.Timestamp(),
				EventType:   events.EventTypeFileOpen,
				EventData:   &ev,
				PID:         ev.Hdr.PID,
				ProcessName: processName,
			})
		}
		return // Skip alert for testing mode
	}

	severity := rule.Severity
	if blocked && severity != "critical" {
		severity = "critical"
	}
	b.emitAlert(apimodel.Alert{
		ID:          fmt.Sprintf("file-%d-%d", ev.Hdr.PID, time.Now().UnixNano()),
		Timestamp:   ev.Hdr.Timestamp().UnixMilli(),
		Severity:    severity,
		RuleName:    rule.Name,
		Description: fmt.Sprintf("%s: %s", rule.Description, filename),
		PID:         ev.Hdr.PID,
		ProcessName: processName,
		CgroupID:    strconv.FormatUint(ev.Hdr.CgroupID, 10),
		Action:      string(rule.Action),
		Blocked:     blocked,
	})
}

func (b *Bridge) HandleConnect(ev events.ConnectEvent) {
	b.mu.RLock()
	re, pt := b.ruleEngine, b.processTree
	b.mu.RUnlock()

	var processName string
	if pt != nil {
		if info, ok := pt.GetProcess(ev.Hdr.PID); ok {
			processName = info.Comm
		}
	}

	b.stats.RecordConnect()
	frontendEvent := ConnectToFrontend(ev, formatAddr(ev), processName)
	b.stats.PublishEvent(frontendEvent)

	if re == nil {
		return
	}

	blocked := ev.Hdr.Blocked == 1

	matched, rule, allowed := re.MatchConnect(&ev)

	// If kernel blocked the connection but Go-side matching failed, still emit alert
	if blocked && (!matched || rule == nil) {
		b.emitAlert(apimodel.Alert{
			ID:          fmt.Sprintf("net-%d-%d", ev.Hdr.PID, time.Now().UnixNano()),
			Timestamp:   ev.Hdr.Timestamp().UnixMilli(),
			Severity:    "critical",
			RuleName:    "Kernel Blocked Connection",
			Description: fmt.Sprintf("Network connection blocked by kernel: %s", formatAddr(ev)),
			PID:         ev.Hdr.PID,
			ProcessName: processName,
			CgroupID:    strconv.FormatUint(ev.Hdr.CgroupID, 10),
			Action:      "block",
			Blocked:     true,
		})
		return
	}

	if !matched || rule == nil || allowed {
		return
	}

	// Handle testing mode: record but don't alert
		if rule.IsTesting() {
		testingBuffer := re.GetTestingBuffer()
		if testingBuffer != nil {
			testingBuffer.RecordHit(&rules.TestingHit{
				RuleName:    rule.Name,
				HitTime:     ev.Hdr.Timestamp(),
				EventType:   events.EventTypeConnect,
				EventData:   &ev,
				PID:         ev.Hdr.PID,
				ProcessName: processName,
			})
		}
		return // Skip alert for testing mode
	}

	severity := rule.Severity
	if blocked && severity != "critical" {
		severity = "critical"
	}
	b.emitAlert(apimodel.Alert{
		ID:          fmt.Sprintf("net-%d-%d", ev.Hdr.PID, time.Now().UnixNano()),
		Timestamp:   ev.Hdr.Timestamp().UnixMilli(),
		Severity:    severity,
		RuleName:    rule.Name,
		Description: rule.Description,
		PID:         ev.Hdr.PID,
		ProcessName: processName,
		CgroupID:    strconv.FormatUint(ev.Hdr.CgroupID, 10),
		Action:      string(rule.Action),
		Blocked:     blocked,
	})
}

func (b *Bridge) emitAlert(alert apimodel.Alert) {
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
