package types

type ExecEvent struct {
	Type       string `json:"type"`
	Timestamp  int64  `json:"timestamp"`
	PID        uint32 `json:"pid"`
	PPID       uint32 `json:"ppid"`
	CgroupID   string `json:"cgroupId"`
	Comm       string `json:"comm"`
	ParentComm string `json:"parentComm"`
	Blocked    bool   `json:"blocked"`
}

type FileEvent struct {
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
	PID       uint32 `json:"pid"`
	CgroupID  string `json:"cgroupId"`
	Flags     uint32 `json:"flags"`
	Ino       uint64 `json:"ino,omitempty"`
	Dev       uint64 `json:"dev,omitempty"`
	Filename  string `json:"filename"`
	Blocked   bool   `json:"blocked"`
}

type ConnectEvent struct {
	Type        string `json:"type"`
	Timestamp   int64  `json:"timestamp"`
	PID         uint32 `json:"pid"`
	ProcessName string `json:"processName,omitempty"`
	CgroupID    string `json:"cgroupId"`
	Family      uint16 `json:"family"`
	Port        uint16 `json:"port"`
	Addr        string `json:"addr"`
	Blocked     bool   `json:"blocked"`
}

type Alert struct {
	ID          string `json:"id"`
	Timestamp   int64  `json:"timestamp"`
	Severity    string `json:"severity"`
	RuleName    string `json:"ruleName"`
	Description string `json:"description"`
	PID         uint32 `json:"pid"`
	ProcessName string `json:"processName"`
	ParentName  string `json:"parentName"`
	CgroupID    string `json:"cgroupId"`
	Action      string `json:"action"`
	Blocked     bool   `json:"blocked"`
}

type Workload struct {
	ID           string `json:"id"`
	CgroupPath   string `json:"cgroupPath"`
	ExecCount    int64  `json:"execCount"`
	FileCount    int64  `json:"fileCount"`
	ConnectCount int64  `json:"connectCount"`
	AlertCount   int64  `json:"alertCount"`
	BlockedCount int64  `json:"blockedCount"`
	FirstSeen    int64  `json:"firstSeen"`
	LastSeen     int64  `json:"lastSeen"`
}

type ProcessInfo struct {
	PID       uint32 `json:"pid"`
	PPID      uint32 `json:"ppid"`
	Comm      string `json:"comm"`
	CgroupID  string `json:"cgroupId"`
	Timestamp int64  `json:"timestamp"`
}
