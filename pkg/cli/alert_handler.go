package cli

import (
	"eulerguard/pkg/events"
	"eulerguard/pkg/output"
	"eulerguard/pkg/proc"
	"eulerguard/pkg/rules"
)

type cliAlertHandler struct {
	processTree *proc.ProcessTree
	printer     *output.Printer
	ruleEngine  *rules.Engine
}

var _ events.EventHandler = (*cliAlertHandler)(nil)

func (h *cliAlertHandler) HandleExec(ev events.ExecEvent) {
	processed := h.printer.Print(ev)
	if _, _, allowed := h.ruleEngine.MatchExec(processed); allowed {
		return
	}
	for _, alert := range h.ruleEngine.CollectExecAlerts(processed) {
		h.printer.PrintAlert(alert)
	}
}

func (h *cliAlertHandler) HandleFileOpen(ev events.FileOpenEvent, filename string) {
	matched, rule, allowed := h.ruleEngine.MatchFile(filename, ev.PID, ev.CgroupID)
	if !matched || rule == nil || allowed {
		return
	}
	h.printer.PrintFileOpenAlert(&ev, h.processTree.GetAncestors(ev.PID), rule, filename)
}

func (h *cliAlertHandler) HandleConnect(ev events.ConnectEvent) {
	matched, rule, allowed := h.ruleEngine.MatchConnect(&ev)
	if !matched || rule == nil || allowed {
		return
	}
	h.printer.PrintConnectAlert(&ev, h.processTree.GetAncestors(ev.PID), rule)
}
