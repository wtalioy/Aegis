package output

import (
	"eulerguard/pkg/events"
	"eulerguard/pkg/utils"
	"strings"
	"time"
)

func (p *Printer) Print(ev events.ExecEvent) events.ProcessedEvent {
	meta := events.ProcessedEvent{
		Event:     ev,
		Timestamp: time.Now().UTC(),
		Process:   normalizeCommand(utils.ExtractCString(ev.Comm[:])),
		Parent:    normalizeCommand(utils.ExtractCString(ev.PComm[:])),
		Rate:      p.meter.Tick(),
	}

	if p.jsonLines {
		p.writeJSON(meta, "exec event")
		return meta
	}

	p.writeLine("[%s] Process executed: PID=%d(%s) ← PPID=%d(%s) | Cgroup=%d | %.1f ev/s\n",
		meta.Timestamp.Format(time.RFC3339),
		meta.Event.PID, meta.Process,
		meta.Event.PPID, meta.Parent,
		meta.Event.CgroupID,
		meta.Rate)

	return meta
}

func normalizeCommand(raw string) string {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "unknown"
	}
	return name
}
