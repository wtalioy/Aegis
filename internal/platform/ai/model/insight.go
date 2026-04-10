package model

import "time"

type InsightType string

const (
	InsightTypeTestingPromotion InsightType = "testing_promotion"
	InsightTypeAnomaly          InsightType = "anomaly"
	InsightTypeOptimization     InsightType = "optimization"
	InsightTypeDailyReport      InsightType = "daily_report"
)

type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type Insight struct {
	ID         string         `json:"id"`
	Type       InsightType    `json:"type"`
	Title      string         `json:"title"`
	Summary    string         `json:"summary"`
	Confidence float64        `json:"confidence"`
	Severity   Severity       `json:"severity"`
	Data       map[string]any `json:"data"`
	Actions    any            `json:"actions"`
	CreatedAt  time.Time      `json:"created_at"`
}
