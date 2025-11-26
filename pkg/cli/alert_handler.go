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
	for _, alert := range h.ruleEngine.Match(processed) {
		h.printer.PrintAlert(alert)
	}
}

func (h *cliAlertHandler) HandleFileOpen(ev events.FileOpenEvent, filename string) {
	if matched, rule := h.ruleEngine.MatchFile(filename, ev.PID, ev.CgroupID); matched && rule != nil {
		chain := h.processTree.GetAncestors(ev.PID)
		h.printer.PrintFileOpenAlert(&ev, chain, rule, filename)
	}
}

func (h *cliAlertHandler) HandleConnect(ev events.ConnectEvent) {
	if matched, rule := h.ruleEngine.MatchConnect(&ev); matched && rule != nil {
		chain := h.processTree.GetAncestors(ev.PID)
		h.printer.PrintConnectAlert(&ev, chain, rule)
	}
}
