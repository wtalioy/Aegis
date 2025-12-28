package types

import (
	"aegis/pkg/rules"
	"aegis/pkg/storage"
)

type Message struct {
	Role      string `json:"role"` // "user", "assistant", "system"
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

type ChatStreamToken struct {
	Content   string `json:"content"`
	Done      bool   `json:"done"`
	SessionID string `json:"sessionId,omitempty"`
	Error     string `json:"error,omitempty"`
}

// DiagnosisResult holds the output of a system diagnosis.
type DiagnosisResult struct {
	Analysis        string `json:"analysis"`
	SnapshotSummary string `json:"snapshotSummary"`
	Provider        string `json:"provider"`
	IsLocal         bool   `json:"isLocal"`
	DurationMs      int64  `json:"durationMs"`
	Timestamp       int64  `json:"timestamp"`
}

type ChatResponse struct {
	Message        string `json:"message"`
	SessionID      string `json:"sessionId"`
	ContextSummary string `json:"contextSummary"`
	Provider       string `json:"provider"`
	IsLocal        bool   `json:"isLocal"`
	DurationMs     int64  `json:"durationMs"`
	Timestamp      int64  `json:"timestamp"`
	MessageCount   int    `json:"messageCount"`
}

const (
	AnalyzeTypeProcess  = "process"
	AnalyzeTypeWorkload = "workload"
	AnalyzeTypeRule     = "rule"
)

type AnalyzeRequest struct {
	Type string `json:"type"` // AnalyzeTypeProcess, AnalyzeTypeWorkload, AnalyzeTypeRule
	ID   string `json:"id"`   // PID, CgroupID, RuleName
}

type Anomaly struct {
	Type        string   `json:"type"`        // "behavior_change", "unusual_pattern", etc.
	Description string   `json:"description"` // Description of the anomaly
	Severity    string   `json:"severity"`    // "low", "medium", "high", "critical"
	Confidence  float64  `json:"confidence"`  // 0-1
	Evidence    []string `json:"evidence"`    // Supporting evidence
}

type Recommendation struct {
	Type        string `json:"type"`        // "rule_creation", "investigation", "baseline_update"
	Description string `json:"description"` // Description of the recommendation
	Priority    string `json:"priority"`    // "low", "medium", "high"
	Action      Action `json:"action"`      // Suggested action
}

type RelatedInsight struct {
	Type    string `json:"type"`    // "correlation", "pattern", "trend"
	Title   string `json:"title"`   // Insight title
	Summary string `json:"summary"` // Insight summary
}

type AnalyzeResponse struct {
	Summary         string           `json:"summary"`
	Anomalies       []Anomaly        `json:"anomalies"`
	BaselineStatus  string           `json:"baseline_status"`
	Recommendations []Recommendation `json:"recommendations"`
	RelatedInsights []RelatedInsight `json:"related_insights"`
}

type ExplainRequest struct {
	EventID   string         `json:"event_id"`
	EventData *storage.Event `json:"event_data"`
	Question  string         `json:"question"`
}

type Action struct {
	Label    string         `json:"label"`
	ActionID string         `json:"action_id"`
	Params   map[string]any `json:"params"`
}

type ExplainResponse struct {
	Explanation      string           `json:"explanation"`
	RootCause        string           `json:"root_cause"`
	MatchedRule      *rules.Rule      `json:"matched_rule"`
	RelatedEvents    []*storage.Event `json:"related_events"`
	SuggestedActions []Action         `json:"suggested_actions"`
}

type RuleGenRequest struct {
	Description string       `json:"description"`
	Examples    []rules.Rule `json:"examples"`
}

type RuleGenResponse struct {
	Rule       rules.Rule `json:"rule"`
	YAML       string     `json:"yaml"`
	Reasoning  string     `json:"reasoning"`
	Confidence float64    `json:"confidence"`
	Warnings   []string   `json:"warnings"`
}

type StatusDTO struct {
	Provider string `json:"provider"`
	IsLocal  bool   `json:"isLocal"`
	Status   string `json:"status"` // "ready", "unavailable"
}
