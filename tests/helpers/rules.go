package helpers

import "aegis/internal/policy"

func ActiveFileRule(name, filename string, action policy.ActionType) policy.Rule {
	return policy.Rule{
		Name:        name,
		Description: name,
		Severity:    "medium",
		Action:      action,
		Type:        policy.RuleTypeFile,
		State:       policy.RuleStateProduction,
		Match: policy.MatchCondition{
			Filename: filename,
		},
	}
}
