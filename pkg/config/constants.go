package config

import "time"

// Storage constants (Phase 1 will use these)
const (
	// DefaultRecentEventsCapacity is the default capacity for recent events storage.
	// Phase 1: TimeRingBuffer capacity (10000+ events)
	DefaultRecentEventsCapacity = 10000

	// DefaultMaxAlerts is the default maximum number of alerts to keep in memory.
	DefaultMaxAlerts = 100

	// DefaultAlertDedupWindow is the default deduplication window for alerts.
	DefaultAlertDedupWindow = 10 * time.Second
)

// Storage configuration (for Phase 1)
type StorageConfig struct {
	RingBufferCapacity int `yaml:"ring_buffer_capacity"` // Default: 10000
}
