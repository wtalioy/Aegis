package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

func LoadRules(filePath string) ([]Rule, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	var ruleSet RuleSet
	if err := yaml.Unmarshal(data, &ruleSet); err != nil {
		return nil, fmt.Errorf("failed to parse rules YAML: %w", err)
	}

	if len(ruleSet.Rules) == 0 {
		return nil, fmt.Errorf("no rules found in file")
	}

	for i := range ruleSet.Rules {
		if ruleSet.Rules[i].Type == "" {
			ruleSet.Rules[i].Type = ruleSet.Rules[i].DeriveType()
		}
	}

	if errs := ValidateRules(ruleSet.Rules); len(errs) > 0 {
		var b strings.Builder
		b.WriteString("rule validation failed:\n")
		for _, err := range errs {
			b.WriteString(" - ")
			b.WriteString(err.Error())
			b.WriteByte('\n')
		}
		return nil, fmt.Errorf("%s", strings.TrimSpace(b.String()))
	}

	return ruleSet.Rules, nil
}

func CleanRuleForYAML(rule Rule) Rule {
	clean := rule
	// Clear metadata fields that shouldn't be in YAML
	clean.CreatedAt = time.Time{}
	clean.DeployedAt = nil
	clean.PromotedAt = nil
	clean.ActualTestingHits = 0
	clean.PromotionScore = 0
	clean.PromotionReasons = nil
	clean.LastReviewedAt = nil
	clean.ReviewNotes = ""
	return clean
}

func SaveRules(filePath string, ruleList []Rule) error {
	// Clean rules before saving (remove metadata fields)
	cleanRules := make([]Rule, len(ruleList))
	for i, rule := range ruleList {
		cleanRules[i] = CleanRuleForYAML(rule)
	}
	
	ruleSet := RuleSet{
		Rules: cleanRules,
	}

	dir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(dir, ".rules-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	encoder := yaml.NewEncoder(tmpFile)
	encoder.SetIndent(2)
	if err := encoder.Encode(ruleSet); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to encode rules to YAML: %w", err)
	}
	if err := encoder.Close(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close encoder: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	if syscall.Geteuid() == 0 {
		if stat, err := os.Stat(filepath.Dir(filePath)); err == nil {
			uid := int(stat.Sys().(*syscall.Stat_t).Uid)
			gid := int(stat.Sys().(*syscall.Stat_t).Gid)
			os.Chown(filePath, uid, gid)
		}
	}

	return nil
}

func MergeRules(existing []Rule, newRules []Rule) []Rule {
	existingSet := make(map[string]bool)
	for _, r := range existing {
		sig := ruleSignature(r)
		existingSet[sig] = true
	}

	result := make([]Rule, len(existing))
	copy(result, existing)

	for _, r := range newRules {
		sig := ruleSignature(r)
		if !existingSet[sig] {
			result = append(result, r)
			existingSet[sig] = true
		}
	}

	return result
}

func ruleSignature(r Rule) string {
	return fmt.Sprintf("%s|%s|%s|%s|%d|%s",
		r.Match.ProcessName,
		r.Match.ParentName,
		r.Match.Filename,
		r.Match.DestIP,
		r.Match.DestPort,
		r.Action,
	)
}

func ValidateRules(rules []Rule) []error {
	var errs []error
	for idx := range rules {
		rule := rules[idx]
		name := strings.TrimSpace(rule.Name)
		displayName := ruleDisplayName(name, idx)

		if name == "" {
			errs = append(errs, fmt.Errorf("rule %d: missing name", idx+1))
		}

		if !isValidAction(rule.Action) {
			errs = append(errs, fmt.Errorf("%s: action must be one of allow, alert, block", displayName))
		}

		switch rule.DeriveType() {
		case RuleTypeExec:
			if !hasExecCondition(rule.Match) {
				errs = append(errs, fmt.Errorf("%s: exec rules require process_name, parent_name, cgroup_id, pid, or ppid", displayName))
			}
		case RuleTypeFile:
			if strings.TrimSpace(rule.Match.Filename) == "" {
				errs = append(errs, fmt.Errorf("%s: file rules require filename", displayName))
			}
		case RuleTypeConnect:
			if rule.Match.DestPort == 0 && strings.TrimSpace(rule.Match.DestIP) == "" && strings.TrimSpace(rule.Match.ProcessName) == "" {
				errs = append(errs, fmt.Errorf("%s: connect rules require dest_port, dest_ip, or process_name", displayName))
			}
		}
	}
	return errs
}

func hasExecCondition(match MatchCondition) bool {
	return strings.TrimSpace(match.ProcessName) != "" ||
		strings.TrimSpace(match.ParentName) != "" ||
		strings.TrimSpace(match.CgroupID) != "" ||
		match.PID != 0 ||
		match.PPID != 0
}

func isValidAction(action ActionType) bool {
	return action == ActionAllow || action == ActionAlert || action == ActionBlock
}

func ruleDisplayName(name string, idx int) string {
	if name != "" {
		return fmt.Sprintf("rule %q", name)
	}
	return fmt.Sprintf("rule #%d", idx+1)
}
