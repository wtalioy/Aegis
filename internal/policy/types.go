package policy

import "aegis/internal/policy/rules"

type MatchType = rules.MatchType
type ActionType = rules.ActionType
type RuleType = rules.RuleType
type RuleState = rules.RuleState
type InodeKey = rules.InodeKey
type MatchCondition = rules.MatchCondition
type Rule = rules.Rule
type RuleSet = rules.RuleSet
type Engine = rules.Engine
type TestingHit = rules.TestingHit
type TestingBuffer = rules.TestingBuffer
type TestingStats = rules.TestingStats
type PromotionReadiness = rules.PromotionReadiness

const (
	ActionAllow ActionType = rules.ActionAllow
	ActionAlert ActionType = rules.ActionAlert
	ActionBlock ActionType = rules.ActionBlock
)

const (
	BPFActionMonitor uint8 = rules.BPFActionMonitor
	BPFActionBlock   uint8 = rules.BPFActionBlock
)

const (
	MatchTypeExact    MatchType = rules.MatchTypeExact
	MatchTypeContains MatchType = rules.MatchTypeContains
	MatchTypePrefix   MatchType = rules.MatchTypePrefix
)

const (
	RuleTypeExec    RuleType = rules.RuleTypeExec
	RuleTypeFile    RuleType = rules.RuleTypeFile
	RuleTypeConnect RuleType = rules.RuleTypeConnect
)

const (
	RuleStateDraft      RuleState = rules.RuleStateDraft
	RuleStateTesting    RuleState = rules.RuleStateTesting
	RuleStateProduction RuleState = rules.RuleStateProduction
	RuleStateArchived   RuleState = rules.RuleStateArchived
)

type TestingRuleStatus struct {
	Rule       Rule               `json:"rule"`
	Validation PromotionReadiness `json:"validation"`
	Stats      TestingStats       `json:"stats"`
}

func CleanRuleForYAML(rule Rule) Rule {
	return rules.CleanRuleForYAML(rule)
}

func NewEngine(ruleList []Rule) *Engine {
	return rules.NewEngine(ruleList)
}

func ValidateRules(ruleList []Rule) []error {
	return rules.ValidateRules(ruleList)
}
