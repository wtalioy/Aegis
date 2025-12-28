package server

import (
	"fmt"
	"time"

	"aegis/pkg/apimodel"
	"aegis/pkg/rules"

	"gopkg.in/yaml.v3"
)

func (a *App) GetSystemStats() SystemStatsDTO {
	processCount := 0
	if a.core != nil && a.core.ProcessTree != nil {
		processCount = a.core.ProcessTree.Size()
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

func (a *App) GetAlerts() []apimodel.Alert {
	return a.stats.Alerts()
}

func (a *App) GetRules() []RuleDTO {
	ruleList := a.GetRulesInternal()
	result := make([]RuleDTO, len(ruleList))

	for i, rule := range ruleList {
		matchMap := buildMatchMap(rule)
		// Clean rule for YAML (remove metadata fields)
		cleanRule := rules.CleanRuleForYAML(rule)
		yamlBytes, _ := yaml.Marshal(cleanRule)

		result[i] = RuleDTO{
			Name:        rule.Name,
			Description: rule.Description,
			Severity:    rule.Severity,
			Action:      string(rule.Action),
			Type:        string(rule.DeriveType()),
			Match:       matchMap,
			YAML:        string(yamlBytes),
			State:       string(rule.State),
			CreatedAt:   &rule.CreatedAt,
			DeployedAt:  rule.DeployedAt,
			PromotedAt:  rule.PromotedAt,
		}
	}

	return result
}

func (a *App) PromoteRule(ruleName string) error {
	if a.core == nil || a.core.RuleEngine == nil {
		return fmt.Errorf("rule engine not available")
	}

	rule, allRules := a.FindRuleByName(ruleName)
	if rule == nil {
		return fmt.Errorf("rule %s not found", ruleName)
	}

	// Update state to production
	rule.State = rules.RuleStateProduction
	now := time.Now()
	rule.PromotedAt = &now

	return a.SaveAndReloadRules(allRules)
}

func (a *App) GetRulesInternal() []rules.Rule {
	if a.core == nil || a.core.RuleEngine == nil {
		return []rules.Rule{}
	}
	return a.core.RuleEngine.GetRules()
}

func (a *App) GetTestingBuffer() *rules.TestingBuffer {
	if a.core == nil || a.core.RuleEngine == nil {
		return nil
	}
	return a.core.RuleEngine.GetTestingBuffer()
}

func (a *App) FindRuleByName(ruleName string) (*rules.Rule, []rules.Rule) {
	allRules := a.GetRulesInternal()
	for i := range allRules {
		if allRules[i].Name == ruleName {
			return &allRules[i], allRules
		}
	}
	return nil, allRules
}

func (a *App) SaveAndReloadRules(allRules []rules.Rule) error {
	if err := rules.SaveRules(a.opts.RulesPath, allRules); err != nil {
		return fmt.Errorf("failed to save rules: %w", err)
	}
	return a.reloadRules()
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
