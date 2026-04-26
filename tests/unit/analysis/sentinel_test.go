package analysis_test

import (
	"testing"
	"time"

	"aegis/internal/analysis/sentinel"
	aiservice "aegis/internal/platform/ai/service"
	"aegis/internal/platform/events"
	"aegis/internal/platform/storage"
	"aegis/internal/policy"
	"aegis/internal/policy/rules"
	"aegis/internal/telemetry/proc"
	"aegis/tests/fakes"
)

func TestSentinelPublishesTestingPromotionInsight(t *testing.T) {
	provider := fakes.NewAIProvider()
	service := aiservice.NewClient(provider)
	engine := rules.NewEngine([]policy.Rule{
		{
			Name:        "testing-shell",
			Description: "testing-shell",
			Severity:    "warning",
			Action:      policy.ActionAlert,
			Type:        policy.RuleTypeExec,
			State:       policy.RuleStateTesting,
			Match: policy.MatchCondition{
				ProcessName: "bash",
			},
		},
	})

	buffer := engine.GetTestingBuffer()
	now := time.Now()
	for i := 0; i < 10; i++ {
		buffer.RecordHit(&policy.TestingHit{
			RuleName:    "testing-shell",
			HitTime:     now.Add(-2 * time.Hour).Add(time.Duration(i) * 15 * time.Minute),
			EventType:   events.EventTypeExec,
			ProcessName: "bash",
		})
	}

	snt := sentinel.NewSentinel(service, engine, storage.NewManager(10, 10), proc.NewProfileRegistry()).
		WithSchedule(sentinel.ScheduleConfig{
			TestingPromotion: 10 * time.Millisecond,
			Anomaly:          time.Hour,
			RuleOptimization: time.Hour,
			DailyReport:      time.Hour,
		})
	defer snt.Stop()
	snt.Start()

	deadline := time.Now().Add(250 * time.Millisecond)
	for time.Now().Before(deadline) {
		insights := snt.GetInsights(10)
		for _, insight := range insights {
			if insight.Type == sentinel.InsightTypeTestingPromotion {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("expected testing promotion insight to be generated")
}
