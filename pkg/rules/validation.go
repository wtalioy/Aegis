package rules

import (
	"fmt"
	"time"

)

type PromotionReadiness struct {
	Score              float64    `json:"score"`                          // 0-1
	IsReady            bool       `json:"is_ready"`                       // Score >= 0.7
	Reasons            []string   `json:"reasons"`                        // Explanation
	MissingCriteria    []string   `json:"missing_criteria"`               // What's needed to be ready
	EstimatedReadyTime *time.Time `json:"estimated_ready_time,omitempty"` // When it might be ready
}

type ValidationService struct {
	testingBuffer                 *TestingBuffer
	promotionMinObservationMinutes int
	promotionMinHits               int
}

func NewValidationService(testingBuffer *TestingBuffer, minObservationMinutes int, minHits int) *ValidationService {
	return &ValidationService{
		testingBuffer:                 testingBuffer,
		promotionMinObservationMinutes: minObservationMinutes,
		promotionMinHits:               minHits,
	}
}

func (vs *ValidationService) CalculatePromotionReadiness(rule *Rule) PromotionReadiness {
	readiness := PromotionReadiness{
		Reasons:         make([]string, 0),
		MissingCriteria: make([]string, 0),
	}

	// Check if rule is in testing state
	if !rule.IsTesting() {
		readiness.IsReady = false
		readiness.MissingCriteria = append(readiness.MissingCriteria, "Rule must be in testing mode")
		return readiness
	}

	// Get testing stats
	stats := vs.testingBuffer.GetStats(rule.Name)

	// Promotion criteria: All must be met (configurable)
	// 1. Observation time
	hasObservationTime := stats.ObservationMinutes >= vs.promotionMinObservationMinutes
	observationHours := float64(stats.ObservationMinutes) / 60.0
	minObservationHours := float64(vs.promotionMinObservationMinutes) / 60.0
	if hasObservationTime {
		readiness.Reasons = append(readiness.Reasons, fmt.Sprintf("Observed for %.1f hours", observationHours))
	} else {
		readiness.MissingCriteria = append(readiness.MissingCriteria,
			fmt.Sprintf("Need %.1f hours observation (currently %.1f hours)", minObservationHours, observationHours))
	}

	// 2. Hit count
	hasEnoughHits := stats.Hits >= vs.promotionMinHits
	if hasEnoughHits {
		readiness.Reasons = append(readiness.Reasons, fmt.Sprintf("Detected %d hits", stats.Hits))
	} else {
		readiness.MissingCriteria = append(readiness.MissingCriteria,
			fmt.Sprintf("Need %d+ hits (currently %d)", vs.promotionMinHits, stats.Hits))
	}

	// Rule is ready if all criteria are met
	readiness.IsReady = hasObservationTime && hasEnoughHits

	// Calculate score: percentage of criteria met (observation time and hits)
	criteriaMet := 0
	totalCriteria := 2
	if hasObservationTime {
		criteriaMet++
	}
	if hasEnoughHits {
		criteriaMet++
	}
	readiness.Score = float64(criteriaMet) / float64(totalCriteria)

	// Estimate when ready
	if !readiness.IsReady {
		readiness.EstimatedReadyTime = vs.calculateEstimatedReadyTime(stats, rule)
	}

	return readiness
}

func (vs *ValidationService) calculateEstimatedReadyTime(stats TestingStats, rule *Rule) *time.Time {
	// Estimate based on current hit rate
	if stats.Hits == 0 {
		return nil
	}

	observationHours := float64(stats.ObservationMinutes) / 60.0
	hitsPerHour := float64(stats.Hits) / observationHours
	hitsNeeded := vs.promotionMinHits - stats.Hits

	if hitsNeeded <= 0 {
		hitsNeeded = 0 // Already have enough hits, check other criteria
	}

	// Estimate time needed for observation (in hours)
	minObservationHours := float64(vs.promotionMinObservationMinutes) / 60.0
	observationNeeded := minObservationHours - observationHours
	if observationNeeded < 0 {
		observationNeeded = 0
	}

	// Take the maximum of both
	hoursNeeded := observationNeeded
	if hitsPerHour > 0 && hitsNeeded > 0 {
		hitsHoursNeeded := float64(hitsNeeded) / hitsPerHour
		if hitsHoursNeeded > hoursNeeded {
			hoursNeeded = hitsHoursNeeded
		}
	}

	if hoursNeeded <= 0 {
		return nil
	}

	estimated := time.Now().Add(time.Duration(hoursNeeded) * time.Hour)
	return &estimated
}
