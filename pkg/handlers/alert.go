package handlers

import (
	"eulerguard/pkg/events"
	"eulerguard/pkg/output"
	"eulerguard/pkg/proctree"
	"eulerguard/pkg/rules"
)

type AlertHandler struct {
	processTree *proctree.ProcessTree
	printer     *output.Printer
	ruleEngine  *rules.Engine
}

var _ events.EventHandler = (*AlertHandler)(nil)

func NewAlertHandler(
	processTree *proctree.ProcessTree,
	printer *output.Printer,
	ruleEngine *rules.Engine,
) *AlertHandler {
	return &AlertHandler{
		processTree: processTree,
		printer:     printer,
		ruleEngine:  ruleEngine,
	}
}

func (h *AlertHandler) HandleExec(ev events.ExecEvent) {
	processedEvent := h.printer.Print(ev)
	for _, alert := range h.ruleEngine.Match(processedEvent) {
		h.printer.PrintAlert(alert)
	}
}

func (h *AlertHandler) HandleFileOpen(ev events.FileOpenEvent, filename string) {
	if matched, rule := h.ruleEngine.MatchFile(filename, ev.PID, ev.CgroupID); matched && rule != nil {
		chain := h.processTree.GetAncestors(ev.PID)
		h.printer.PrintFileOpenAlert(&ev, chain, rule, filename)
	}
}

func (h *AlertHandler) HandleConnect(ev events.ConnectEvent) {
	if matched, rule := h.ruleEngine.MatchConnect(&ev); matched && rule != nil {
		chain := h.processTree.GetAncestors(ev.PID)
		h.printer.PrintConnectAlert(&ev, chain, rule)
	}
}

