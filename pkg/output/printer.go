package output

import (
	"bufio"
	"encoding/json"
	"eulerguard/pkg/events"
	"eulerguard/pkg/metrics"
	"eulerguard/pkg/proctree"
	"eulerguard/pkg/rules"
	"eulerguard/pkg/utils"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type threadSafeWriter struct {
	mu     *sync.Mutex
	writer io.Writer
}

func (w *threadSafeWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(p)
}

type Printer struct {
	jsonLines  bool
	meter      *metrics.RateMeter
	logFile    *os.File
	logWriter  *bufio.Writer
	writer     io.Writer
	flushTimer *time.Ticker
	stopFlush  chan struct{}
	closeOnce  sync.Once
	mu         *sync.Mutex
}

func NewPrinter(jsonLines bool, meter *metrics.RateMeter, logPath string, bufferSize int) (*Printer, error) {
	if err := rotateLogIfNeeded(logPath); err != nil {
		log.Printf("Warning: log rotation failed: %v", err)
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logWriter := bufio.NewWriterSize(f, bufferSize)
	mu := &sync.Mutex{}
	tsLogWriter := &threadSafeWriter{
		mu:     mu,
		writer: logWriter,
	}

	p := &Printer{
		jsonLines:  jsonLines,
		meter:      meter,
		logFile:    f,
		logWriter:  logWriter,
		writer:     io.MultiWriter(os.Stdout, tsLogWriter),
		flushTimer: time.NewTicker(1 * time.Second),
		stopFlush:  make(chan struct{}),
		mu:         mu,
	}

	// Start periodic flush goroutine
	go func() {
		for {
			select {
			case <-p.flushTimer.C:
				p.mu.Lock()
				if p.logWriter != nil {
					_ = p.logWriter.Flush()
				}
				p.mu.Unlock()
			case <-p.stopFlush:
				return
			}
		}
	}()

	log.Printf("Logging to file: %s", logPath)
	return p, nil
}

func (p *Printer) Close() error {
	var closeErr error

	p.closeOnce.Do(func() {
		// Stop flush goroutine
		if p.flushTimer != nil {
			p.flushTimer.Stop()
		}
		if p.stopFlush != nil {
			close(p.stopFlush)
		}

		// Flush remaining data (protected by mutex)
		if p.logWriter != nil {
			p.mu.Lock()
			if err := p.logWriter.Flush(); err != nil {
				log.Printf("Warning: failed to flush log buffer: %v", err)
			}
			p.mu.Unlock()
		}

		if p.logFile != nil {
			closeErr = p.logFile.Close()
		}
	})

	return closeErr
}

func (p *Printer) Print(ev events.ExecEvent) events.ProcessedEvent {
	processName := strings.TrimSpace(utils.ExtractCString(ev.Comm[:]))
	if processName == "" {
		processName = "unknown"
	}

	parentName := strings.TrimSpace(utils.ExtractCString(ev.PComm[:]))
	if parentName == "" {
		parentName = "unknown"
	}

	meta := events.ProcessedEvent{
		Event:     ev,
		Timestamp: time.Now().UTC(),
		Process:   processName,
		Parent:    parentName,
		Rate:      p.meter.Tick(),
	}

	if p.jsonLines {
		enc := json.NewEncoder(p.writer)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(meta); err != nil {
			log.Printf("json encode failed: %v", err)
		}
		return meta
	}

	fmt.Fprintf(p.writer, "[%s] Process executed: PID=%d(%s) ← PPID=%d(%s) | Cgroup=%d | %.1f ev/s\n",
		meta.Timestamp.Format(time.RFC3339),
		meta.Event.PID, meta.Process,
		meta.Event.PPID, meta.Parent,
		meta.Event.CgroupID,
		meta.Rate)

	return meta
}

func (p *Printer) PrintAlert(alert rules.Alert) {
	if p.jsonLines {
		alertData := map[string]any{
			"type":        "alert",
			"timestamp":   alert.Event.Timestamp.Format(time.RFC3339),
			"rule_name":   alert.Rule.Name,
			"severity":    alert.Rule.Severity,
			"description": alert.Message,
			"pid":         alert.Event.Event.PID,
			"process":     alert.Event.Process,
			"ppid":        alert.Event.Event.PPID,
			"parent":      alert.Event.Parent,
			"cgroup_id":   alert.Event.Event.CgroupID,
		}
		enc := json.NewEncoder(p.writer)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(alertData); err != nil {
			log.Printf("json encode alert failed: %v", err)
		}
		return
	}

	alertText := formatAlertText(alert.Rule.Name, alert.Rule.Severity, alert.Message,
		alert.Event.Event.PID, alert.Event.Process,
		alert.Event.Event.PPID, alert.Event.Parent,
		alert.Event.Event.CgroupID)

	severityColor := getSeverityColor(alert.Rule.Severity)
	resetColor := "\033[0m"
	fmt.Fprintf(os.Stdout, "%s%s%s", severityColor, alertText, resetColor)

	p.mu.Lock()
	fmt.Fprint(p.logWriter, alertText)
	p.mu.Unlock()
}

func formatAlertText(ruleName, severity, description string, pid uint32, process string, ppid uint32, parent string, cgroupID uint64) string {
	return fmt.Sprintf("[Alert!] Rule '%s' triggered [Severity: %s]\n"+
		"  Description: %s\n"+
		"  Process: PID=%d(%s) ← PPID=%d(%s) | Cgroup=%d\n",
		ruleName, severity, description,
		pid, process, ppid, parent, cgroupID)
}

func (p *Printer) PrintFileOpenAlert(ev *events.FileOpenEvent, chain []*proctree.ProcessInfo, rule *rules.Rule, filename string) {
	if p.jsonLines {
		data := map[string]any{
			"type":        "file_access_alert",
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"rule_name":   rule.Name,
			"severity":    rule.Severity,
			"description": rule.Description,
			"pid":         ev.PID,
			"filename":    filename,
			"cgroup_id":   ev.CgroupID,
			"flags":       ev.Flags,
			"chain":       formatChainJSON(chain),
		}
		enc := json.NewEncoder(p.writer)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(data); err != nil {
			log.Printf("json encode file alert failed: %v", err)
		}
		return
	}

	alertText := formatFileAlertText(rule.Name, rule.Severity, rule.Description,
		filename, ev.PID, ev.CgroupID, ev.Flags, chain)

	severityColor := getSeverityColor(rule.Severity)
	resetColor := "\033[0m"
	fmt.Fprintf(os.Stdout, "%s%s%s", severityColor, alertText, resetColor)

	p.mu.Lock()
	fmt.Fprint(p.logWriter, alertText)
	p.mu.Unlock()
}

func formatFileAlertText(ruleName, severity, description, filename string,
	pid uint32, cgroupID uint64, flags uint32, chain []*proctree.ProcessInfo) string {

	var builder strings.Builder
	fmt.Fprintf(&builder, "[ALERT!] Rule '%s' triggered [Severity: %s]\n", ruleName, severity)
	fmt.Fprintf(&builder, "  Description: %s\n", description)
	fmt.Fprintf(&builder, "  File: %s\n", filename)
	fmt.Fprintf(&builder, "  PID: %d | Cgroup: %d | Flags: 0x%x\n", pid, cgroupID, flags)

	if len(chain) > 0 {
		fmt.Fprintf(&builder, "  Attack Chain: %s\n", formatChain(chain))
	}

	return builder.String()
}

func (p *Printer) PrintConnectAlert(ev *events.ConnectEvent, chain []*proctree.ProcessInfo, rule *rules.Rule) {
	destAddr := formatAddress(ev)

	if p.jsonLines {
		data := map[string]any{
			"type":        "network_connect_alert",
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"rule_name":   rule.Name,
			"severity":    rule.Severity,
			"description": rule.Description,
			"pid":         ev.PID,
			"dest_addr":   destAddr,
			"dest_port":   ev.Port,
			"family":      ev.Family,
			"cgroup_id":   ev.CgroupID,
			"chain":       formatChainJSON(chain),
		}
		enc := json.NewEncoder(p.writer)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(data); err != nil {
			log.Printf("json encode connect alert failed: %v", err)
		}
		return
	}

	alertText := formatConnectAlertText(rule.Name, rule.Severity, rule.Description,
		destAddr, ev.PID, ev.CgroupID, chain)

	severityColor := getSeverityColor(rule.Severity)
	resetColor := "\033[0m"
	fmt.Fprintf(os.Stdout, "%s%s%s", severityColor, alertText, resetColor)

	p.mu.Lock()
	fmt.Fprint(p.logWriter, alertText)
	p.mu.Unlock()
}

func formatConnectAlertText(ruleName, severity, description, destAddr string,
	pid uint32, cgroupID uint64, chain []*proctree.ProcessInfo) string {

	var builder strings.Builder
	fmt.Fprintf(&builder, "[ALERT!] Rule '%s' triggered [Severity: %s]\n", ruleName, severity)
	fmt.Fprintf(&builder, "  Description: %s\n", description)
	fmt.Fprintf(&builder, "  Network Connection: %s\n", destAddr)
	fmt.Fprintf(&builder, "  PID: %d | Cgroup: %d\n", pid, cgroupID)

	if len(chain) > 0 {
		fmt.Fprintf(&builder, "  Attack Chain: %s\n", formatChain(chain))
	}

	return builder.String()
}

func formatAddress(ev *events.ConnectEvent) string {
	ip := utils.ExtractIP(ev)
	switch ev.Family {
	case 10:
		return fmt.Sprintf("[%s]:%d", ip, ev.Port)
	case 2:
		return fmt.Sprintf("%s:%d", ip, ev.Port)
	default:
		return fmt.Sprintf("unknown_family_%d:port_%d", ev.Family, ev.Port)
	}
}

func formatChain(chain []*proctree.ProcessInfo) string {
	parts := make([]string, len(chain))
	for i, info := range reverseChain(chain) {
		parts[i] = fmt.Sprintf("%s(%d)", info.Comm, info.PID)
	}
	return strings.Join(parts, " -> ")
}

func formatChainJSON(chain []*proctree.ProcessInfo) []map[string]any {
	result := make([]map[string]any, len(chain))
	for i, info := range reverseChain(chain) {
		result[i] = map[string]any{
			"pid":       info.PID,
			"ppid":      info.PPID,
			"comm":      info.Comm,
			"cgroup_id": info.CgroupID,
		}
	}
	return result
}

func reverseChain(chain []*proctree.ProcessInfo) []*proctree.ProcessInfo {
	reversed := make([]*proctree.ProcessInfo, len(chain))
	for i, info := range chain {
		reversed[len(chain)-1-i] = info
	}
	return reversed
}

func getSeverityColor(severity string) string {
	switch severity {
	case "high", "critical":
		return "\033[1;31m" // Bold Red
	case "warning", "medium":
		return "\033[1;33m" // Bold Yellow
	case "info", "low":
		return "\033[1;36m" // Bold Cyan
	default:
		return "\033[1;37m" // Bold White
	}
}
