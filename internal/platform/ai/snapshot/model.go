package snapshot

import "time"

type SystemState struct {
	Timestamp     time.Time `json:"timestamp"`
	LoadLevel     string    `json:"loadLevel"` // "low", "normal", "high", "critical"
	ExecRate      int64     `json:"execRate"`
	FileRate      int64     `json:"fileRate"`
	NetworkRate   int64     `json:"networkRate"`
	ProcessCount  int       `json:"processCount"`
	WorkloadCount int       `json:"workloadCount"`
	AlertCount    int       `json:"alertCount"`

	TopWorkloads      []WorkloadSummary    `json:"topWorkloads"`
	RecentAlerts      []AlertSummary       `json:"recentAlerts"`
	RecentProcesses   []ProcessActivity    `json:"recentProcesses"`
	RecentConnections []ConnectionActivity `json:"recentConnections"`
	RecentFileAccess  []FileActivity       `json:"recentFileAccess"`
}

type WorkloadSummary struct {
	ID          string `json:"id"`
	CgroupPath  string `json:"cgroupPath"`
	TotalEvents int64  `json:"totalEvents"`
	AlertCount  int64  `json:"alertCount"`
}

type AlertSummary struct {
	RuleName    string `json:"ruleName"`
	Severity    string `json:"severity"`
	ProcessName string `json:"processName"`
	Count       int    `json:"count"`
	WasBlocked  bool   `json:"wasBlocked"`
}

type ProcessActivity struct {
	Comm       string `json:"comm"`
	ParentComm string `json:"parentComm"`
	Count      int    `json:"count"`
	Blocked    bool   `json:"blocked"`
}

type ConnectionActivity struct {
	Destination string `json:"destination"`
	Count       int    `json:"count"`
	Blocked     bool   `json:"blocked"`
}

type FileActivity struct {
	Path    string `json:"path"`
	Count   int    `json:"count"`
	Blocked bool   `json:"blocked"`
}
