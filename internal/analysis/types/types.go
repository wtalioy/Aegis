package types

import (
	"aegis/internal/policy"
	"time"
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
	BaselineStatus  string           `json:"baselineStatus"`
	Recommendations []Recommendation `json:"recommendations"`
	RelatedInsights []RelatedInsight `json:"relatedInsights"`
}

type ExplainRequest struct {
	EventID  string `json:"eventId"`
	Question string `json:"question"`
}

type ActionParams struct {
	RuleName    string `json:"ruleName,omitempty"`
	InsightID   string `json:"insightId,omitempty"`
	Page        string `json:"page,omitempty"`
	EventID     string `json:"eventId,omitempty"`
	ContextType string `json:"contextType,omitempty"`
}

type Action struct {
	Label    string       `json:"label"`
	ActionID string       `json:"actionId"`
	Params   ActionParams `json:"params"`
}

type RelatedEvent struct {
	Type        string `json:"type"`
	Timestamp   int64  `json:"timestamp"`
	PID         uint32 `json:"pid,omitempty"`
	PPID        uint32 `json:"ppid,omitempty"`
	CgroupID    string `json:"cgroupId,omitempty"`
	ProcessName string `json:"processName,omitempty"`
	Filename    string `json:"filename,omitempty"`
	Port        uint16 `json:"port,omitempty"`
	Blocked     bool   `json:"blocked"`
}

type ExplainResponse struct {
	Explanation      string         `json:"explanation"`
	RootCause        string         `json:"rootCause"`
	MatchedRule      *policy.Rule   `json:"matchedRule"`
	RelatedEvents    []RelatedEvent `json:"relatedEvents"`
	SuggestedActions []Action       `json:"suggestedActions"`
}

type RuleGenRequest struct {
	Description string        `json:"description"`
	Examples    []policy.Rule `json:"examples"`
}

type RuleGenResponse struct {
	Rule       policy.Rule `json:"rule"`
	YAML       string      `json:"yaml"`
	Reasoning  string      `json:"reasoning"`
	Confidence float64     `json:"confidence"`
	Warnings   []string    `json:"warnings"`
}

type StatusDTO struct {
	Provider string `json:"provider"`
	IsLocal  bool   `json:"isLocal"`
	Status   string `json:"status"` // "ready", "unavailable"
}

type InsightData struct {
	Kind             string  `json:"kind,omitempty"`
	RuleName         string  `json:"ruleName,omitempty"`
	Hits             int     `json:"hits,omitempty"`
	ObservationHours float64 `json:"observationHours,omitempty"`
	EventCount       int     `json:"eventCount,omitempty"`
	RuleCount        int     `json:"ruleCount,omitempty"`
	Date             string  `json:"date,omitempty"`
	Summary          string  `json:"summary,omitempty"`
}

type Insight struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Title      string      `json:"title"`
	Summary    string      `json:"summary"`
	Severity   string      `json:"severity"`
	Confidence float64     `json:"confidence"`
	CreatedAt  time.Time   `json:"createdAt"`
	Actions    []Action    `json:"actions"`
	Data       InsightData `json:"data"`
}

type AskInsightRequest struct {
	Insight  Insight `json:"insight"`
	Question string  `json:"question"`
}

type AskInsightResponse struct {
	Answer     string  `json:"answer"`
	Confidence float64 `json:"confidence"`
}
