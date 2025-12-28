package ai

import "time"

// SentinelScheduleConfig allows tests or callers to override task intervals.
type SentinelScheduleConfig struct {
	TestingPromotion time.Duration
	Anomaly          time.Duration
	RuleOptimization time.Duration
	DailyReport      time.Duration
}

// defaultSentinelSchedule returns the baked-in production cadence.
func defaultSentinelSchedule() SentinelScheduleConfig {
	return SentinelScheduleConfig{
		TestingPromotion: 5 * time.Minute,
		Anomaly:          1 * time.Minute,
		RuleOptimization: 30 * time.Minute,
		DailyReport:      24 * time.Hour,
	}
}
