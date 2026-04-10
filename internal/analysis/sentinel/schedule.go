package sentinel

import "time"

// ScheduleConfig allows tests or callers to override task intervals.
type ScheduleConfig struct {
	TestingPromotion time.Duration
	Anomaly          time.Duration
	RuleOptimization time.Duration
	DailyReport      time.Duration
}

// defaultSchedule returns the baked-in production cadence.
func defaultSchedule() ScheduleConfig {
	return ScheduleConfig{
		TestingPromotion: 5 * time.Minute,
		Anomaly:          1 * time.Minute,
		RuleOptimization: 30 * time.Minute,
		DailyReport:      24 * time.Hour,
	}
}
