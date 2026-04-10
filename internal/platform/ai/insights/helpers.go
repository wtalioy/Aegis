package insights

import (
	"fmt"
	"time"

	"aegis/internal/analysis/types"
)

func NewInsightID(prefix string, now time.Time) string {
	return fmt.Sprintf("%s-%d", prefix, now.Unix())
}

// NewInsight creates a new Insight.
//
// NOTE: Insight types live in the sentinel package. To avoid a sentinel <-> insights
// import cycle, this helper is defined in insights but expects the caller to pass
// a pointer to a struct compatible with the following shape. In practice, only
// internal analysis sentinel code uses this.
//
// If you want to remove this indirection later, move Insight definition out of
// sentinel into a neutral package.
func NewInsight(now time.Time, id string, typ any, title, summary string, severity any) *struct {
	ID         string         `json:"id"`
	Type       any            `json:"type"`
	Title      string         `json:"title"`
	Summary    string         `json:"summary"`
	Confidence float64        `json:"confidence"`
	Severity   any            `json:"severity"`
	Data       map[string]any `json:"data"`
	Actions    []types.Action `json:"actions"`
	CreatedAt  time.Time      `json:"created_at"`
} {
	return &struct {
		ID         string         `json:"id"`
		Type       any            `json:"type"`
		Title      string         `json:"title"`
		Summary    string         `json:"summary"`
		Confidence float64        `json:"confidence"`
		Severity   any            `json:"severity"`
		Data       map[string]any `json:"data"`
		Actions    []types.Action `json:"actions"`
		CreatedAt  time.Time      `json:"created_at"`
	}{
		ID:         id,
		Type:       typ,
		Title:      title,
		Summary:    summary,
		Confidence: 1.0,
		Severity:   severity,
		Data:       map[string]any{},
		Actions:    []types.Action{},
		CreatedAt:  now,
	}
}
