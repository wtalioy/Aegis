package rules

import (
	"sync"
	"time"

	"aegis/pkg/events"
	"aegis/pkg/storage"
)

type TestingHit struct {
	RuleName      string
	HitTime       time.Time
	EventType     events.EventType
	EventData     any // Can be *events.ExecEvent, *events.FileOpenEvent, or *events.ConnectEvent
	PID           uint32
	ProcessName   string
	FalsePositive bool // Set by AI analysis (Phase 3)
}

// TestingBuffer stores testing mode rule hits.
type TestingBuffer struct {
	mu    sync.RWMutex
	hits  []*TestingHit
	store storage.EventStore // Use TimeRingBuffer for storage
}

func NewTestingBuffer(capacity int) *TestingBuffer {
	return &TestingBuffer{
		hits:  make([]*TestingHit, 0, capacity),
		store: storage.NewTimeRingBuffer(capacity),
	}
}

// RecordHit records a testing mode rule hit.
func (tb *TestingBuffer) RecordHit(hit *TestingHit) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Store in TimeRingBuffer
	event := &storage.Event{
		Type:      hit.EventType,
		Timestamp: hit.HitTime,
		Data:      hit.EventData,
	}
	_ = tb.store.Append(event)

	// Also keep in memory for quick access
	tb.hits = append(tb.hits, hit)

	// Limit memory size (keep last N hits)
	if len(tb.hits) > 10000 {
		tb.hits = tb.hits[len(tb.hits)-10000:]
	}
}

func (tb *TestingBuffer) GetHits(ruleName string, timeWindow time.Duration) []*TestingHit {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	cutoff := time.Now().Add(-timeWindow)
	var result []*TestingHit

	for _, hit := range tb.hits {
		if hit.RuleName == ruleName && hit.HitTime.After(cutoff) {
			result = append(result, hit)
		}
	}

	return result
}

func (tb *TestingBuffer) GetHitsByRule(ruleName string) []*TestingHit {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	var result []*TestingHit
	for _, hit := range tb.hits {
		if hit.RuleName == ruleName {
			result = append(result, hit)
		}
	}

	return result
}

func (tb *TestingBuffer) GetStats(ruleName string) TestingStats {
	hits := tb.GetHitsByRule(ruleName)
	if len(hits) == 0 {
		return TestingStats{
			RuleName: ruleName,
		}
	}

	// Calculate observation time
	oldest := hits[0].HitTime
	newest := hits[0].HitTime
	for _, hit := range hits {
		if hit.HitTime.Before(oldest) {
			oldest = hit.HitTime
		}
		if hit.HitTime.After(newest) {
			newest = hit.HitTime
		}
	}
	observationDuration := newest.Sub(oldest)
	observationMinutes := int(observationDuration.Minutes())

	// Count hits by process
	processCounts := make(map[string]int)
	for _, hit := range hits {
		processCounts[hit.ProcessName]++
	}

	hitsByProcess := make([]ProcessHitCount, 0, len(processCounts))
	for name, count := range processCounts {
		hitsByProcess = append(hitsByProcess, ProcessHitCount{
			Name:  name,
			Count: count,
		})
	}

	return TestingStats{
		RuleName:           ruleName,
		Hits:               len(hits),
		ObservationMinutes: observationMinutes,
		HitsByProcess:      hitsByProcess,
	}
}

type TestingStats struct {
	RuleName           string            `json:"ruleName"`
	Hits               int               `json:"hits"`
	ObservationMinutes int               `json:"observationMinutes"`
	HitsByProcess      []ProcessHitCount `json:"hitsByProcess"`
}

type ProcessHitCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ClearHits clears all hits for a specific rule
func (tb *TestingBuffer) ClearHits(ruleName string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Filter out hits for this rule
	var remaining []*TestingHit
	for _, hit := range tb.hits {
		if hit.RuleName != ruleName {
			remaining = append(remaining, hit)
		}
	}
	tb.hits = remaining
}
