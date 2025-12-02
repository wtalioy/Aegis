package rules

import (
	"testing"

	"eulerguard/pkg/types"
)

func TestValidateRulesValid(t *testing.T) {
	rules := []types.Rule{
		{
			Name:   "Allow bash",
			Action: types.ActionAllow,
			Match: types.MatchCondition{
				ProcessName: "bash",
			},
			Type: types.RuleTypeExec,
		},
		{
			Name:   "Protect shadow",
			Action: types.ActionAlert,
			Match: types.MatchCondition{
				Filename: "/etc/shadow",
			},
			Type: types.RuleTypeFile,
		},
		{
			Name:   "Block port",
			Action: types.ActionBlock,
			Match: types.MatchCondition{
				DestPort: 4444,
			},
			Type: types.RuleTypeConnect,
		},
	}

	if errs := ValidateRules(rules); len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %v", errs)
	}
}

func TestValidateRulesInvalid(t *testing.T) {
	rules := []types.Rule{
		{}, // Missing everything
		{
			Name:   "Bad Connect",
			Action: "drop",
			Type:   types.RuleTypeConnect,
		},
	}

	errs := ValidateRules(rules)
	if len(errs) < 2 {
		t.Fatalf("expected multiple errors, got %v", errs)
	}
}
