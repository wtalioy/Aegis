package events

import "time"

type ExecEvent struct {
	PID      uint32
	PPID     uint32
	CgroupID uint64
	Comm     [16]byte
	PComm    [16]byte
}

type ProcessedEvent struct {
	Event     ExecEvent
	Timestamp time.Time
	Process   string
	Parent    string
	Rate      float64
}
