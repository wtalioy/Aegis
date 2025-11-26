package ui

type FrontendExecEvent struct {
	Type        string `json:"type"`
	Timestamp   int64  `json:"timestamp"`
	PID         uint32 `json:"pid"`
	PPID        uint32 `json:"ppid"`
	CgroupID    string `json:"cgroupId"`
	Comm        string `json:"comm"`
	ParentComm  string `json:"parentComm"`
	InContainer bool   `json:"inContainer"`
}

type FrontendConnectEvent struct {
	Type        string `json:"type"`
	Timestamp   int64  `json:"timestamp"`
	PID         uint32 `json:"pid"`
	CgroupID    string `json:"cgroupId"`
	Family      uint16 `json:"family"`
	Port        uint16 `json:"port"`
	Addr        string `json:"addr"`
	InContainer bool   `json:"inContainer"`
}

type FrontendFileEvent struct {
	Type        string `json:"type"`
	Timestamp   int64  `json:"timestamp"`
	PID         uint32 `json:"pid"`
	CgroupID    string `json:"cgroupId"`
	Flags       uint32 `json:"flags"`
	Filename    string `json:"filename"`
	InContainer bool   `json:"inContainer"`
}

type FrontendAlert struct {
	ID          string `json:"id"`
	Timestamp   int64  `json:"timestamp"`
	Severity    string `json:"severity"`
	RuleName    string `json:"ruleName"`
	Description string `json:"description"`
	PID         uint32 `json:"pid"`
	ProcessName string `json:"processName"`
	ParentName  string `json:"parentName"`
	CgroupID    string `json:"cgroupId"`
	InContainer bool   `json:"inContainer"`
}

type SystemStatsDTO struct {
	ProcessCount   int     `json:"processCount"`
	ContainerCount int     `json:"containerCount"`
	EventsPerSec   float64 `json:"eventsPerSec"`
	AlertCount     int     `json:"alertCount"`
	ProbeStatus    string  `json:"probeStatus"` // "active", "error", "starting"
}

type ProcessInfoDTO struct {
	PID       uint32 `json:"pid"`
	PPID      uint32 `json:"ppid"`
	Comm      string `json:"comm"`
	CgroupID  string `json:"cgroupId"`
	Timestamp int64  `json:"timestamp"`
}

type RuleDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Action      string `json:"action"`
	YAML        string `json:"yaml"`
	Selected    bool   `json:"selected"`
}

type LearningStatusDTO struct {
	Active           bool  `json:"active"`
	StartTime        int64 `json:"startTime"`
	Duration         int   `json:"duration"` // seconds
	PatternCount     int   `json:"patternCount"`
	ExecCount        int   `json:"execCount"`
	FileCount        int   `json:"fileCount"`
	ConnectCount     int   `json:"connectCount"`
	RemainingSeconds int   `json:"remainingSeconds"`
}

// DetectionRuleDTO represents a loaded detection rule for the frontend
type DetectionRuleDTO struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Severity    string            `json:"severity"`
	Action      string            `json:"action"`
	Type        string            `json:"type"` // "exec", "file", "connect"
	Match       map[string]string `json:"match"`
	YAML        string            `json:"yaml"`
}

// ProbeStatsDTO represents real-time probe statistics
type ProbeStatsDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Tracepoint string `json:"tracepoint"`
	Active     bool   `json:"active"`
	EventsRate int64  `json:"eventsRate"`  // events per second
	TotalCount int64  `json:"totalCount"`  // total events captured
}
