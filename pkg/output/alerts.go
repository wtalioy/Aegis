package output

import (
	"eulerguard/pkg/events"
	"eulerguard/pkg/proc"
	"eulerguard/pkg/rules"
	"time"
)

func (p *Printer) PrintAlert(alert rules.Alert) {
	if p.jsonLines {
		p.writeJSON(map[string]any{
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
		}, "exec alert")
		return
	}

	alertText := formatAlertText(alert.Rule.Name, alert.Rule.Severity, alert.Message,
		alert.Event.Event.PID, alert.Event.Process,
		alert.Event.Event.PPID, alert.Event.Parent,
		alert.Event.Event.CgroupID)

	p.emitColoredAlert(alert.Rule.Severity, alertText)
}

func (p *Printer) PrintFileOpenAlert(ev *events.FileOpenEvent, chain []*proc.ProcessInfo, rule *rules.Rule, filename string) {
	if p.jsonLines {
		p.writeJSON(map[string]any{
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
		}, "file alert")
		return
	}

	alertText := formatFileAlertText(rule.Name, rule.Severity, rule.Description,
		filename, ev.PID, ev.CgroupID, ev.Flags, chain)

	p.emitColoredAlert(rule.Severity, alertText)
}

func (p *Printer) PrintConnectAlert(ev *events.ConnectEvent, chain []*proc.ProcessInfo, rule *rules.Rule) {
	destAddr := formatAddress(ev)

	if p.jsonLines {
		p.writeJSON(map[string]any{
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
		}, "connect alert")
		return
	}

	alertText := formatConnectAlertText(rule.Name, rule.Severity, rule.Description,
		destAddr, ev.PID, ev.CgroupID, chain)

	p.emitColoredAlert(rule.Severity, alertText)
}
