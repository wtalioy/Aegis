package events

import "time"

type EventType uint8

const (
	EventTypeExec     EventType = 1
	EventTypeFileOpen EventType = 2
	EventTypeConnect  EventType = 3

	// Buffer sizes (must match BPF definitions)
	TaskCommLen      = 16
	PathMaxLen       = 256
	CommandLineLen   = 512 // Full command line (executable + all args)

	// EventHeaderSize is the size of the unified event header (56 bytes)
	EventHeaderSize = 56
)

type EventHeader struct {
	TimestampNs uint64
	CgroupID    uint64
	PID         uint32
	TID         uint32
	UID         uint32
	GID         uint32
	Type        EventType
	Blocked     uint8
	_           [6]byte // padding
	Comm        [TaskCommLen]byte
}

type ExecEvent struct {
	Hdr         EventHeader
	PPID        uint32
	_           [4]byte // padding
	PComm       [TaskCommLen]byte
	Filename    [PathMaxLen]byte
	CommandLine [CommandLineLen]byte
}

type FileOpenEvent struct {
	Hdr      EventHeader
	Ino      uint64
	Dev      uint64
	Flags    uint32
	_        [4]byte // padding
	Filename [PathMaxLen]byte
}

type ConnectEvent struct {
	Hdr    EventHeader
	AddrV4 uint32
	Family uint16
	Port   uint16
	AddrV6 [16]byte
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
