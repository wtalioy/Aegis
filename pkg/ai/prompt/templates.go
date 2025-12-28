package prompt

type TokenBudget struct {
	SystemPrompt int // 系统提示词预算
	Context      int // 上下文预算
	UserInput    int // 用户输入预算
	Response     int // 响应预算
}

var DefaultBudgets = map[string]TokenBudget{
	"rulegen":  {800, 500, 200, 500}, // 中等复杂度
	"explain":  {600, 800, 100, 600}, // 需要较多上下文
	"sentinel": {400, 1000, 0, 400},  // 数据密集型
}

type PromptContext struct {
	CurrentPage        string
	SelectedItem       string
	RecentActions      []string
	Input              string
	ExistingRules      []string
	RecentBlocked      []string
	TargetWorkload     string
	EventType          string
	ProcessName        string
	PID                uint32
	ParentName         string
	Target             string
	Action             string
	RuleName           string
	ProcessHistory     []EventHistory
	RelatedProcesses   []RelatedProcess
	Type               string
	ID                 string
	CommandLine        string
	StartTime          string
	FileOpenCount      int64
	NetConnectCount    int64
	CgroupPath         string
	ProcessCount       int
	TotalEvents        int64
	TestingRuleName    string // For testing rule analysis
	ObservationMinutes int    // Observation time in minutes
	TotalHits          int
	HitsByProcess      []ProcessHit
	SampleEvents       []SampleEvent
	BaselineFileRate   float64
	BaselineNetRate    float64
	BaselineFiles      []string
	CurrentFileRate    float64
	CurrentNetRate     float64
	UnusualFiles       []string
	UnusualConnections []string
}

type EventHistory struct {
	Timestamp   string
	Description string
}

type RelatedProcess struct {
	Comm       string
	EventCount int
}

type ProcessHit struct {
	ProcessName string
	Count       int
}

type SampleEvent struct {
	Timestamp   string
	ProcessName string
	Target      string
}

const DiagnosisSystemPrompt = `You are Aegis AI, an expert Linux kernel security analyst. 
You analyze eBPF telemetry data to diagnose system issues and security threats.
Be concise, technical, and actionable. Use markdown formatting.
Focus on: root cause, security implications, and remediation steps.`

