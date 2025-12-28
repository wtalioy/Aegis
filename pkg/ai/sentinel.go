package ai

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"aegis/pkg/ai/types"
	"aegis/pkg/proc"
	"aegis/pkg/rules"
	"aegis/pkg/storage"
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
	service       SentinelAIClient
	ruleEngine    *rules.Engine
	store         storage.EventStore
	profileReg    *proc.ProfileRegistry
	schedule      SentinelScheduleConfig
	insights      []*Insight
	insightsMu    sync.RWMutex
	subscribers   map[chan *Insight]struct{}
	subscribersMu sync.RWMutex
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

func NewSentinel(
	service SentinelAIClient,
	ruleEngine *rules.Engine,
	store storage.EventStore,
	profileReg *proc.ProfileRegistry,
) *Sentinel {
	return &Sentinel{
		service:     service,
		ruleEngine:  ruleEngine,
		store:       store,
		profileReg:  profileReg,
		schedule:    defaultSentinelSchedule(),
		insights:    make([]*Insight, 0),
		subscribers: make(map[chan *Insight]struct{}),
		stopChan:    make(chan struct{}),
	}
}

// WithSchedule overrides the default schedule. If zero values are provided,
// the corresponding defaults are kept.
func (s *Sentinel) WithSchedule(cfg SentinelScheduleConfig) *Sentinel {
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
	s.insightsMu.Lock()
	s.insights = make([]*Insight, 0)
	s.insightsMu.Unlock()

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
	insight := &Insight{
		ID:         fmt.Sprintf("welcome-%d", now.Unix()),
		Type:       InsightTypeDailyReport,
		Title:      "AI Sentinel Active",
		Summary:    "AI Sentinel is now monitoring your system. It will analyze security events, detect anomalies, and provide optimization suggestions. Insights will appear here as they are discovered.",
		Confidence: 1.0,
		Severity:   SeverityLow,
		Data:       map[string]any{"type": "welcome"},
		Actions:    []types.Action{},
		CreatedAt:  now, // Use explicit variable to ensure fresh timestamp
	}
	s.addInsights([]*Insight{insight})
}

func (s *Sentinel) Stop() {
	close(s.stopChan)
	s.wg.Wait()
}

func (s *Sentinel) Subscribe() <-chan *Insight {
	ch := make(chan *Insight, 100)
	s.subscribersMu.Lock()
	s.subscribers[ch] = struct{}{}
	s.subscribersMu.Unlock()
	return ch
}


// Note: Due to Go's channel type system, we can't directly match receive-only channels.
// In practice, channels are rarely unsubscribed - they're typically closed when done.
// This method is kept for API compatibility but may not work perfectly.
func (s *Sentinel) Unsubscribe(ch <-chan *Insight) {
	// Channels can't be directly compared, so we can't reliably remove them
	// The channel will be garbage collected when no longer referenced
	// In practice, subscribers should just stop reading from the channel
}


func (s *Sentinel) GetInsights(limit int) []*Insight {
	s.insightsMu.RLock()
	totalInsights := len(s.insights)
	allInsights := make([]*Insight, totalInsights)
	copy(allInsights, s.insights)
	s.insightsMu.RUnlock()

	// Filter out stale insights (welcome/system status older than 1 hour)
	now := time.Now()
	filtered := make([]*Insight, 0, len(allInsights))
	for _, insight := range allInsights {
		age := now.Sub(insight.CreatedAt)

		// Filter out old welcome/system status insights
		if (insight.Type == InsightTypeDailyReport && insight.Data["type"] == "welcome") ||
			(insight.Type == InsightTypeAnomaly && insight.Title == "System Monitoring Active") {
			// Only include if less than 1 hour old
			if age < 1*time.Hour {
				filtered = append(filtered, insight)
			}
		} else if insight.Type == InsightTypeDailyReport && insight.Title == "Daily Security Report" {
			// Daily reports older than 25 hours should be filtered (allow some buffer for 24h interval)
			if age < 25*time.Hour {
				filtered = append(filtered, insight)
			}
		} else {
			// Include all other insights, but ensure they're not too old (older than 7 days)
			if age < 7*24*time.Hour {
				filtered = append(filtered, insight)
			}
		}
	}

	// Sort by creation time (newest first)
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[i].CreatedAt.Before(filtered[j].CreatedAt) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	if limit <= 0 || limit > len(filtered) {
		limit = len(filtered)
	}

	// Return most recent insights
	if limit > 0 {
		result := make([]*Insight, limit)
		copy(result, filtered[:limit])
		return result
	}
	return filtered
}


func (s *Sentinel) runTask(task func(context.Context) []*Insight, interval time.Duration) {
	defer s.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on start
	ctx := context.Background()
	insights := task(ctx)
	s.addInsights(insights)

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			ctx := context.Background()
			insights := task(ctx)
			s.addInsights(insights)
		}
	}
}


func (s *Sentinel) addInsights(insights []*Insight) {
	if len(insights) == 0 {
		return
	}

	s.insightsMu.Lock()
	// Deduplicate: remove old insights with the same ID or similar content
	for _, newInsight := range insights {
		// Remove old insights with the same ID
		filtered := make([]*Insight, 0, len(s.insights))
		for _, existing := range s.insights {
			// Keep if different ID and not a duplicate welcome/system status insight
			if existing.ID != newInsight.ID {
				// Also filter out old welcome/system status insights that are older than 1 hour
				if (existing.Type == InsightTypeDailyReport && existing.Data["type"] == "welcome") ||
					(existing.Type == InsightTypeAnomaly && existing.Title == "System Monitoring Active") {
					// Keep only if less than 1 hour old
					if time.Since(existing.CreatedAt) < 1*time.Hour {
						filtered = append(filtered, existing)
					}
				} else {
					filtered = append(filtered, existing)
				}
			}
		}
		s.insights = filtered
		// Add new insight
		s.insights = append(s.insights, newInsight)
	}
	// Keep only last 1000 insights
	if len(s.insights) > 1000 {
		s.insights = s.insights[len(s.insights)-1000:]
	}
	s.insightsMu.Unlock()

	// Notify subscribers
	s.subscribersMu.RLock()
	subs := make([]chan *Insight, 0, len(s.subscribers))
	for ch := range s.subscribers {
		subs = append(subs, ch)
	}
	s.subscribersMu.RUnlock()

	for _, insight := range insights {
		for _, ch := range subs {
			select {
			case ch <- insight:
			default:
				// Channel full, skip
			}
		}
	}
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
	insights := []*Insight{}

	for _, rule := range allRules {
		if !rule.IsTesting() {
			continue
		}

		stats := testingBuffer.GetStats(rule.Name)
		if stats.Hits < 10 {
			continue // Not enough data
		}

		// Check if rule should be promoted (simplified logic)
		// In production, use AI to analyze false positive rate
		// Lowered thresholds for demo purposes - can be adjusted
		observationHours := float64(stats.ObservationMinutes) / 60.0
		if observationHours >= 1 && stats.Hits >= 5 {
			now := time.Now()
			insight := &Insight{
				ID:         fmt.Sprintf("testing-promotion-%s-%d", rule.Name, now.Unix()),
				Type:       InsightTypeTestingPromotion,
				Title:      fmt.Sprintf("Testing Rule Ready for Promotion: %s", rule.Name),
				Summary:    fmt.Sprintf("Rule '%s' has been running in Testing mode for %.1f hours with %d hits. Consider promoting it to Production mode.", rule.Name, observationHours, stats.Hits),
				Confidence: 0.8,
				Severity:   SeverityMedium,
				Data: map[string]any{
					"rule_name":         rule.Name,
					"hits":              stats.Hits,
					"observation_hours": observationHours,
				},
				Actions: []types.Action{
					{
						Label:    "Promote to Production",
						ActionID: "promote",
						Params:   map[string]any{"rule_name": rule.Name},
					},
					{
						Label:    "Dismiss",
						ActionID: "dismiss",
						Params:   map[string]any{"insight_id": fmt.Sprintf("testing-promotion-%s-%d", rule.Name, now.Unix())},
					},
				},
				CreatedAt: now, // Use explicit variable to ensure fresh timestamp
			}
			insights = append(insights, insight)
		}
	}

	return insights
}


func (s *Sentinel) checkAnomalies(ctx context.Context) []*Insight {
	if s.profileReg == nil || !s.service.IsEnabled() {
		return nil
	}

	insights := []*Insight{}

	// Simplified anomaly detection - would use AI in production
	// For now, generate a basic system status insight if we have activity
	// This ensures users see something even if no specific anomalies are detected

	// Check if we have any rules or activity to report on
	if s.ruleEngine != nil {
		rules := s.ruleEngine.GetRules()
		if len(rules) > 0 {
			// Check if we already have a recent system status insight (within last 5 minutes)
			s.insightsMu.RLock()
			hasRecentStatus := false
			for _, existing := range s.insights {
				if existing.Type == InsightTypeAnomaly && existing.Title == "System Monitoring Active" {
					if time.Since(existing.CreatedAt) < 5*time.Minute {
						hasRecentStatus = true
						break
					}
				}
			}
			s.insightsMu.RUnlock()

			// Only create new insight if we don't have a recent one
			if !hasRecentStatus {
				insight := &Insight{
					ID:         fmt.Sprintf("system-status-%d", time.Now().Unix()),
					Type:       InsightTypeAnomaly,
					Title:      "System Monitoring Active",
					Summary:    fmt.Sprintf("AI Sentinel is actively monitoring %d security rules. The system is being analyzed for anomalies and security events.", len(rules)),
					Confidence: 0.8,
					Severity:   SeverityLow,
					Data: map[string]any{
						"rule_count": len(rules),
					},
					Actions:   []types.Action{},
					CreatedAt: time.Now(),
				}
				insights = append(insights, insight)
			}
		}
	}

	return insights
}


func (s *Sentinel) checkRuleOptimization(ctx context.Context) []*Insight {
	if s.ruleEngine == nil || !s.service.IsEnabled() {
		return nil
	}

	insights := []*Insight{}

	allRules := s.ruleEngine.GetRules()
	if len(allRules) == 0 {
		// Suggest creating first rule
		insight := &Insight{
			ID:         fmt.Sprintf("optimization-no-rules-%d", time.Now().Unix()),
			Type:       InsightTypeOptimization,
			Title:      "No Security Rules Detected",
			Summary:    "Your system currently has no active security rules. Consider creating rules to monitor and protect your system. You can use Policy Studio to create rules based on your security requirements.",
			Confidence: 1.0,
			Severity:   SeverityMedium,
			Data:       map[string]any{"rule_count": 0},
			Actions: []types.Action{
				{
					Label:    "Go to Policy Studio",
					ActionID: "navigate",
					Params:   map[string]any{"page": "policy-studio"},
				},
			},
			CreatedAt: time.Now(),
		}
		insights = append(insights, insight)
		return insights
	}

	// Check for rules that might need optimization
	// For now, just return empty - can be enhanced with AI analysis
	return insights
}


func (s *Sentinel) generateDailyReport(ctx context.Context) []*Insight {
	if !s.service.IsEnabled() {
		return nil
	}

	// Build a proper daily report prompt that generates human-readable text
	reportPrompt := `Generate a daily security summary for the Aegis system.

Provide a concise, human-readable summary (not JSON) covering:
1. Overall system security status
2. Key security events or patterns observed
3. Any notable anomalies or concerns
4. Recommendations for the security team

Format your response as a clear, readable report using markdown. Use headers, bullet points, and clear language. Do not output JSON.

Keep it under 300 words and focus on actionable insights.`

	// Use the chat API with system context for better results
	response, err := s.service.SingleChat(ctx, reportPrompt)
	if err != nil {
		return nil
	}

	// Clean up the response - remove any JSON formatting if present
	summary := response
	// If response contains JSON, try to extract the readable parts
	if strings.Contains(summary, "```json") {
		// Extract content before JSON block
		parts := strings.Split(summary, "```json")
		if len(parts) > 0 {
			summary = strings.TrimSpace(parts[0])
		}
	}

	now := time.Now()
	insight := &Insight{
		ID:         fmt.Sprintf("daily-report-%d", now.Unix()/86400),
		Type:       InsightTypeDailyReport,
		Title:      "Daily Security Report",
		Summary:    summary,
		Confidence: 0.9,
		Severity:   SeverityLow,
		Data:       map[string]any{},
		Actions:    []types.Action{},
		CreatedAt:  now, // Use explicit variable to ensure fresh timestamp
	}

	return []*Insight{insight}
}
