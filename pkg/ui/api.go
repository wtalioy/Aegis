package ui

import (
	"fmt"
	"strconv"
	"time"

	"eulerguard/pkg/profiler"
	"eulerguard/pkg/rules"

	"gopkg.in/yaml.v3"
)

func (a *App) GetSystemStats() SystemStatsDTO {
	processCount := 0
	if a.processTree != nil {
		processCount = a.processTree.Size()
	}

	exec, file, net := a.stats.Rates()

	return SystemStatsDTO{
		ProcessCount:  processCount,
		WorkloadCount: a.stats.WorkloadCount(),
		EventsPerSec:  float64(exec + file + net),
		AlertCount:    int(a.stats.TotalAlertCount()),
		ProbeStatus:   "active",
	}
}

func (a *App) GetAncestors(pid uint32) []ProcessInfoDTO {
	if a.processTree == nil {
		return nil
	}

	chain := a.processTree.GetAncestors(pid)
	result := make([]ProcessInfoDTO, len(chain))

	for i, info := range chain {
		result[i] = ProcessInfoDTO{
			PID:       info.PID,
			PPID:      info.PPID,
			Comm:      info.Comm,
			CgroupID:  strconv.FormatUint(info.CgroupID, 10),
			Timestamp: info.Timestamp.UnixMilli(),
		}
	}
	return result
}

func (a *App) GetAlerts() []FrontendAlert {
	return a.stats.Alerts()
}

func (a *App) GetLearningStatus() LearningStatusDTO {
	if a.profiler == nil {
		return LearningStatusDTO{Active: false}
	}

	active := a.learning.active && a.profiler.IsActive()
	var remaining int
	if active {
		elapsed := time.Since(a.learning.startTime)
		remaining = int((a.learning.duration - elapsed).Seconds())
		if remaining < 0 {
			remaining = 0
		}
	}

	execCount, fileCount, connectCount := a.profiler.Counts()

	return LearningStatusDTO{
		Active:           active,
		StartTime:        a.learning.startTime.UnixMilli(),
		Duration:         int(a.learning.duration.Seconds()),
		PatternCount:     a.profiler.Count(),
		ExecCount:        execCount,
		FileCount:        fileCount,
		ConnectCount:     connectCount,
		RemainingSeconds: remaining,
	}
}

func (a *App) StartLearning(durationSec int) error {
	if a.learning.active {
		return fmt.Errorf("learning mode already active")
	}

	a.profiler = profiler.NewProfiler()
	a.learning.active = true
	a.learning.startTime = time.Now()
	a.learning.duration = time.Duration(durationSec) * time.Second

	a.bridge.SetProfiler(a.profiler)

	go func() {
		time.Sleep(a.learning.duration)
		if a.learning.active {
			a.profiler.Stop()
			a.learning.active = false
		}
	}()

	return nil
}

func (a *App) StopLearning() ([]RuleDTO, error) {
	if a.profiler == nil {
		return nil, fmt.Errorf("no profiler data available")
	}

	if a.learning.active {
		a.profiler.Stop()
		a.learning.active = false
	}

	a.bridge.SetProfiler(nil)

	generatedRules := a.profiler.GenerateRules()
	result := make([]RuleDTO, len(generatedRules))

	for i, rule := range generatedRules {
		yamlBytes, _ := yaml.Marshal(rule)
		matchMap := buildMatchMap(rule)
		result[i] = RuleDTO{
			Name:        rule.Name,
			Description: rule.Description,
			Severity:    rule.Severity,
			Action:      string(rule.Action),
			Type:        string(rule.Type),
			Match:       matchMap,
			YAML:        string(yamlBytes),
			Selected:    true,
		}
	}
	return result, nil
}

func (a *App) ApplyWhitelistRules(ruleIndices []int) error {
	if a.profiler == nil {
		return fmt.Errorf("no profiler data available")
	}

	selectedRules := a.profiler.GenerateRulesFiltered(ruleIndices)
	if len(selectedRules) == 0 {
		return fmt.Errorf("no rules selected")
	}

	existingRules, err := rules.LoadRules(a.opts.RulesPath)
	if err != nil {
		existingRules = []rules.Rule{}
	}

	mergedRules := rules.MergeRules(existingRules, selectedRules)

	if err := rules.SaveRules(a.opts.RulesPath, mergedRules); err != nil {
		return fmt.Errorf("failed to save rules: %w", err)
	}

	return nil
}

func (a *App) GetProbeInfo() []map[string]string {
	return []map[string]string{
		{"id": "exec", "name": "Process Execution", "tracepoint": "tp/sched/sched_process_exec"},
		{"id": "openat", "name": "File Access", "tracepoint": "tp/syscalls/sys_enter_openat"},
		{"id": "connect", "name": "Network Connection", "tracepoint": "tp/syscalls/sys_enter_connect"},
	}
}

func (a *App) GetProbeStats() []ProbeStatsDTO {
	execRate, fileRate, netRate := a.stats.Rates()
	execCount, fileCount, netCount := a.stats.Counts()

	return []ProbeStatsDTO{
		{
			ID:         "exec",
			Name:       "Process Execution",
			Tracepoint: "tp/sched/sched_process_exec",
			Active:     true,
			EventsRate: execRate,
			TotalCount: execCount,
		},
		{
			ID:         "openat",
			Name:       "File Access",
			Tracepoint: "tp/syscalls/sys_enter_openat",
			Active:     true,
			EventsRate: fileRate,
			TotalCount: fileCount,
		},
		{
			ID:         "connect",
			Name:       "Network Connection",
			Tracepoint: "tp/syscalls/sys_enter_connect",
			Active:     true,
			EventsRate: netRate,
			TotalCount: netCount,
		},
	}
}

func (a *App) GetWorkloads() []WorkloadDTO {
	if a.workloadRegistry == nil {
		return []WorkloadDTO{}
	}

	workloads := a.workloadRegistry.List()
	result := make([]WorkloadDTO, len(workloads))

	for i, w := range workloads {
		result[i] = WorkloadDTO{
			ID:           strconv.FormatUint(uint64(w.ID), 10),
			CgroupPath:   w.CgroupPath,
			ExecCount:    w.ExecCount,
			FileCount:    w.FileCount,
			ConnectCount: w.ConnectCount,
			AlertCount:   w.AlertCount,
			FirstSeen:    w.FirstSeen.UnixMilli(),
			LastSeen:     w.LastSeen.UnixMilli(),
		}
	}

	return result
}

func (a *App) GetWorkload(id string) *WorkloadDTO {
	if a.workloadRegistry == nil {
		return nil
	}

	cgroupID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil
	}

	w := a.workloadRegistry.Get(cgroupID)
	if w == nil {
		return nil
	}

	return &WorkloadDTO{
		ID:           strconv.FormatUint(uint64(w.ID), 10),
		CgroupPath:   w.CgroupPath,
		ExecCount:    w.ExecCount,
		FileCount:    w.FileCount,
		ConnectCount: w.ConnectCount,
		AlertCount:   w.AlertCount,
		FirstSeen:    w.FirstSeen.UnixMilli(),
		LastSeen:     w.LastSeen.UnixMilli(),
	}
}

func (a *App) GetRules() []RuleDTO {
	if a.ruleEngine == nil {
		return []RuleDTO{}
	}

	ruleList := a.ruleEngine.GetRules()
	result := make([]RuleDTO, len(ruleList))

	for i, rule := range ruleList {
		matchMap := buildMatchMap(rule)
		yamlBytes, _ := yaml.Marshal(rule)

		result[i] = RuleDTO{
			Name:        rule.Name,
			Description: rule.Description,
			Severity:    rule.Severity,
			Action:      string(rule.Action),
			Type:        string(rule.DeriveType()),
			Match:       matchMap,
			YAML:        string(yamlBytes),
		}
	}

	return result
}

func buildMatchMap(rule rules.Rule) map[string]string {
	matchMap := make(map[string]string)
	if rule.Match.ProcessName != "" {
		matchMap["process_name"] = rule.Match.ProcessName
	}
	if rule.Match.ParentName != "" {
		matchMap["parent_name"] = rule.Match.ParentName
	}
	if rule.Match.Filename != "" {
		matchMap["filename"] = rule.Match.Filename
	}
	if rule.Match.FilePath != "" {
		matchMap["file_path"] = rule.Match.FilePath
	}
	if rule.Match.DestPort != 0 {
		matchMap["dest_port"] = fmt.Sprintf("%d", rule.Match.DestPort)
	}
	if rule.Match.DestIP != "" {
		matchMap["dest_ip"] = rule.Match.DestIP
	}
	if rule.Match.CgroupID != "" {
		matchMap["cgroup_id"] = rule.Match.CgroupID
	}
	return matchMap
}
