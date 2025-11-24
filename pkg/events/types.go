package events

import "time"

type EventType uint8

const (
	EventTypeExec     EventType = 1
	EventTypeFileOpen EventType = 2
	EventTypeConnect  EventType = 3

	// Buffer sizes (must match BPF definitions)
	TaskCommLen = 16
	PathMaxLen  = 256
)

type ExecEvent struct {
	PID      uint32
	PPID     uint32
	CgroupID uint64
	Comm     [TaskCommLen]byte
	PComm    [TaskCommLen]byte
}

type FileOpenEvent struct {
	PID      uint32
	CgroupID uint64
	Flags    uint32
	Filename [PathMaxLen]byte
}

type ConnectEvent struct {
	PID      uint32
	CgroupID uint64
	Family   uint16 // AF_INET or AF_INET6
	Port     uint16 // destination port in host byte order
	AddrV4   uint32 // IPv4 address for AF_INET
	AddrV6   [16]byte // IPv6 address for AF_INET6
}

type Event struct {
	Type     EventType
	Exec     *ExecEvent
	FileOpen *FileOpenEvent
}

type ProcessedEvent struct {
	Event     ExecEvent
	Timestamp time.Time
	Process   string
	Parent    string
	Rate      float64
}
