package server

import "time"

type SystemStatsDTO struct {
	ProcessCount  int     `json:"processCount"`
	WorkloadCount int     `json:"workloadCount"`
	EventsPerSec  float64 `json:"eventsPerSec"`
	AlertCount    int     `json:"alertCount"`
	ProbeStatus   string  `json:"probeStatus"` // "active", "error", "starting"
}

type RuleDTO struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Severity    string            `json:"severity"`
	Action      string            `json:"action"`
	Type        string            `json:"type"` // "exec", "file", "connect"
	State       string            `json:"state,omitempty"` // "production", "testing", "draft", "archived"
	Match       map[string]string `json:"match,omitempty"`
	YAML        string            `json:"yaml"`
	Selected    bool              `json:"selected,omitempty"`
	CreatedAt   *time.Time        `json:"created_at,omitempty"`
	DeployedAt  *time.Time        `json:"deployed_at,omitempty"`
	PromotedAt  *time.Time        `json:"promoted_at,omitempty"`
}

