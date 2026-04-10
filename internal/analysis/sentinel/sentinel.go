package sentinel

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"aegis/internal/analysis/types"
	"aegis/internal/platform/ai/insights"
	"aegis/internal/platform/storage"
	"aegis/internal/policy"
	"aegis/internal/telemetry/proc"
)

// SentinelAIClient is the minimal AI capability Sentinel depends on.
// It is intentionally smaller than the full Service surface to keep
// background monitoring decoupled from higher-level features.
type SentinelAIClient interface {
	IsEnabled() bool
	SingleChat(ctx context.Context, prompt string) (string, error)
}

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

// Insight represents a single Sentinel insight item.
type Insight struct {
	ID         string         `json:"id"`
	Type       InsightType    `json:"type"`
	Title      string         `json:"title"`
	Summary    string         `json:"summary"`
	Confidence float64        `json:"confidence"`
	Severity   Severity       `json:"severity"`
	Data       map[string]any `json:"data"`
	Actions    []types.Action `json:"actions"`
	CreatedAt  time.Time      `json:"created_at"`
}

type Sentinel struct {
	service    SentinelAIClient
	ruleEngine *policy.Engine
	store      storage.EventStore
	profileReg *proc.ProfileRegistry
	schedule   ScheduleConfig

	insights *insights.Store[*Insight]

	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewSentinel(
	service SentinelAIClient,
	ruleEngine *policy.Engine,
	store storage.EventStore,
	profileReg *proc.ProfileRegistry,
) *Sentinel {
	return &Sentinel{
		service:    service,
		ruleEngine: ruleEngine,
		store:      store,
		profileReg: profileReg,
		schedule:   defaultSchedule(),
		insights:   insights.NewStore[*Insight](),
		stopChan:   make(chan struct{}),
	}
}

// WithSchedule overrides the default schedule. If zero values are provided,
// the corresponding defaults are kept.
func (s *Sentinel) WithSchedule(cfg ScheduleConfig) *Sentinel {
	if cfg.TestingPromotion != 0 {
		s.schedule.TestingPromotion = cfg.TestingPromotion
	}
	if cfg.Anomaly != 0 {
		s.schedule.Anomaly = cfg.Anomaly
	}
	if cfg.RuleOptimization != 0 {
		s.schedule.RuleOptimization = cfg.RuleOptimization
	}
	if cfg.DailyReport != 0 {
		s.schedule.DailyReport = cfg.DailyReport
	}
	return s
}

func (s *Sentinel) Start() {
	// Clear any old insights when starting fresh
	s.insights.Reset()

	// Generate initial welcome insight with fresh timestamp
	s.generateWelcomeInsight()

	s.wg.Add(4)
	go s.runTask(s.checkTestingPromotion, s.schedule.TestingPromotion)
	go s.runTask(s.checkAnomalies, s.schedule.Anomaly)
	go s.runTask(s.checkRuleOptimization, s.schedule.RuleOptimization)
	go s.runTask(s.generateDailyReport, s.schedule.DailyReport)
}

func (s *Sentinel) generateWelcomeInsight() {
	now := time.Now()
	raw := insights.NewInsight(
		now,
		insights.NewInsightID("welcome", now),
		InsightTypeDailyReport,
		"AI Sentinel Active",
		"AI Sentinel is now monitoring your system. It will analyze security events, detect anomalies, and provide optimization suggestions. Insights will appear here as they are discovered.",
		SeverityLow,
	)
	insight := &Insight{
		ID:         raw.ID,
		Type:       raw.Type.(InsightType),
		Title:      raw.Title,
		Summary:    raw.Summary,
		Confidence: raw.Confidence,
		Severity:   raw.Severity.(Severity),
		Data:       raw.Data,
		Actions:    raw.Actions,
		CreatedAt:  raw.CreatedAt,
	}
	insight.Data["type"] = "welcome"
	s.addInsights([]*Insight{insight})
}

func (s *Sentinel) Stop() {
	close(s.stopChan)
	s.wg.Wait()
}

func (s *Sentinel) Subscribe() insights.Subscription[*Insight] {
	return s.insights.Subscribe(100)
}

// Unsubscribe is kept for API compatibility with older callers.
// Prefer using the Subscription.Cancel() returned by Subscribe().

func (s *Sentinel) GetInsights(limit int) []*Insight {
	return s.insights.List(limit)
}

func (s *Sentinel) runTask(task func(context.Context) []*Insight, interval time.Duration) {
	defer s.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on start
	ctx := context.Background()
	s.addInsights(task(ctx))

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			ctx := context.Background()
			s.addInsights(task(ctx))
		}
	}
}

func (s *Sentinel) addInsights(insights []*Insight) {
	s.insights.Add(insights, func(a, b *Insight) bool {
		// newest first
		return a.CreatedAt.After(b.CreatedAt)
	})
}

func (s *Sentinel) checkTestingPromotion(ctx context.Context) []*Insight {
	if s.ruleEngine == nil || !s.service.IsEnabled() {
		return nil
	}

	testingBuffer := s.ruleEngine.GetTestingBuffer()
	if testingBuffer == nil {
		return nil
	}

	allRules := s.ruleEngine.GetRules()
	out := make([]*Insight, 0)

	for _, rule := range allRules {
		if !rule.IsTesting() {
			continue
		}

		stats := testingBuffer.GetStats(rule.Name)
		if stats.Hits < 10 {
			continue // Not enough data
		}

		observationHours := float64(stats.ObservationMinutes) / 60.0
		if observationHours >= 1 && stats.Hits >= 5 {
			now := time.Now()
			id := fmt.Sprintf("testing-promotion-%s-%d", rule.Name, now.Unix())
			raw := insights.NewInsight(
				now,
				id,
				InsightTypeTestingPromotion,
				fmt.Sprintf("Testing Rule Ready for Promotion: %s", rule.Name),
				fmt.Sprintf("Rule '%s' has been running in Testing mode for %.1f hours with %d hits. Consider promoting it to Production mode.", rule.Name, observationHours, stats.Hits),
				SeverityMedium,
			)
			insight := &Insight{
				ID:         raw.ID,
				Type:       raw.Type.(InsightType),
				Title:      raw.Title,
				Summary:    raw.Summary,
				Confidence: raw.Confidence,
				Severity:   raw.Severity.(Severity),
				Data:       raw.Data,
				Actions:    raw.Actions,
				CreatedAt:  raw.CreatedAt,
			}
			insight.Confidence = 0.8
			insight.Data["rule_name"] = rule.Name
			insight.Data["hits"] = stats.Hits
			insight.Data["observation_hours"] = observationHours
			insight.Actions = []types.Action{
				{Label: "Promote to Production", ActionID: "promote", Params: map[string]any{"rule_name": rule.Name}},
				{Label: "Dismiss", ActionID: "dismiss", Params: map[string]any{"insight_id": id}},
			}
			out = append(out, insight)
		}
	}

	return out
}

func (s *Sentinel) checkAnomalies(ctx context.Context) []*Insight {
	if s.store == nil || !s.service.IsEnabled() {
		return nil
	}

	// Check for recent notable events in the last 15 minutes
	events, err := s.store.Query(time.Now().Add(-15*time.Minute), time.Now())
	if err != nil || len(events) == 0 {
		// No events, so we can create a "Normal" insight
		now := time.Now()
		raw := insights.NewInsight(
			now,
			insights.NewInsightID("system-status-normal", now),
			InsightTypeAnomaly,
			"System Status: Normal",
			"No notable security events detected in the last 15 minutes. System is operating as expected.",
			SeverityLow,
		)
		insight := &Insight{
			ID:         raw.ID,
			Type:       raw.Type.(InsightType),
			Title:      raw.Title,
			Summary:    raw.Summary,
			Confidence: 0.95,
			Severity:   raw.Severity.(Severity),
			Data:       raw.Data,
			Actions:    raw.Actions,
			CreatedAt:  raw.CreatedAt,
		}
		return []*Insight{insight}
	}

	// If we have events, create a higher-severity insight
	now := time.Now()
	raw := insights.NewInsight(
		now,
		insights.NewInsightID("suspicious-activity", now),
		InsightTypeAnomaly,
		"Suspicious Activity Detected",
		fmt.Sprintf("Detected %d notable security events in the last 15 minutes that may require investigation.", len(events)),
		SeverityMedium,
	)
	insight := &Insight{
		ID:         raw.ID,
		Type:       raw.Type.(InsightType),
		Title:      raw.Title,
		Summary:    raw.Summary,
		Confidence: 0.8,
		Severity:   raw.Severity.(Severity),
		Data:       raw.Data,
		Actions:    raw.Actions,
		CreatedAt:  raw.CreatedAt,
	}
	insight.Data["event_count"] = len(events)
	insight.Actions = []types.Action{{Label: "Investigate Events", ActionID: "navigate", Params: map[string]any{"page": "observatory"}}}
	return []*Insight{insight}
}

func (s *Sentinel) checkRuleOptimization(ctx context.Context) []*Insight {
	if s.ruleEngine == nil || !s.service.IsEnabled() {
		return nil
	}

	allRules := s.ruleEngine.GetRules()
	if len(allRules) != 0 {
		return nil
	}

	now := time.Now()
	raw := insights.NewInsight(
		now,
		insights.NewInsightID("optimization-no-rules", now),
		InsightTypeOptimization,
		"No Security Rules Detected",
		"Your system currently has no active security rules. Consider creating rules to monitor and protect your system. You can use Policy Studio to create rules based on your security requirements.",
		SeverityMedium,
	)
	insight := &Insight{
		ID:         raw.ID,
		Type:       raw.Type.(InsightType),
		Title:      raw.Title,
		Summary:    raw.Summary,
		Confidence: raw.Confidence,
		Severity:   raw.Severity.(Severity),
		Data:       raw.Data,
		Actions:    raw.Actions,
		CreatedAt:  raw.CreatedAt,
	}
	insight.Actions = []types.Action{{Label: "Go to Policy Studio", ActionID: "navigate", Params: map[string]any{"page": "policy-studio"}}}
	insight.Data["rule_count"] = 0
	return []*Insight{insight}
}

func (s *Sentinel) generateDailyReport(ctx context.Context) []*Insight {
	if !s.service.IsEnabled() {
		return nil
	}

	// Gather context for the report
	recentInsights := s.insights.List(10)
	var insightSummary strings.Builder
	if len(recentInsights) > 1 { // More than just the welcome message
		insightSummary.WriteString("Here is a summary of recent activity to analyze:")
		for _, insight := range recentInsights {
			// Skip the welcome message and old daily reports
			if insight.Data["type"] == "welcome" || insight.Type == InsightTypeDailyReport {
				continue
			}
			insightSummary.WriteString(fmt.Sprintf("- At %s, this insight was generated: '%s' with summary: '%s'\n", insight.CreatedAt.Format(time.RFC1123), insight.Title, insight.Summary))
		}
	}

	reportPrompt := fmt.Sprintf(`Generate a daily security summary for the system.

Provide a concise, human-readable summary (not JSON) covering:
1. Overall system security status
2. Key security events or patterns observed (based on the provided context)
3. Any notable anomalies or concerns
4. Recommendations

%s

Format your response as a direct, professional report using markdown. Use headers and bullet points. Do not add any conversational filler or a concluding summary sentence. If no significant events occurred, state that the system is operating normally.`, insightSummary.String())

	response, err := s.service.SingleChat(ctx, reportPrompt)
	if err != nil {
		// Log error
		return nil
	}

	summary := response
	if strings.Contains(summary, "```json") {
		parts := strings.Split(summary, "```json")
		if len(parts) > 0 {
			summary = strings.TrimSpace(parts[0])
		}
	}

	now := time.Now()
	id := fmt.Sprintf("daily-report-%d", now.Unix()/86400)
	raw := insights.NewInsight(
		now,
		id,
		InsightTypeDailyReport,
		"Daily Security Report",
		summary,
		SeverityLow,
	)
	insight := &Insight{
		ID:         raw.ID,
		Type:       raw.Type.(InsightType),
		Title:      raw.Title,
		Summary:    raw.Summary,
		Confidence: 0.9,
		Severity:   raw.Severity.(Severity),
		Data:       raw.Data,
		Actions:    raw.Actions,
		CreatedAt:  raw.CreatedAt,
	}
	return []*Insight{insight}
}
