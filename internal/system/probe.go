package system

const (
	ProbeStatusStarting = "starting"
	ProbeStatusActive   = "active"
	ProbeStatusError    = "error"
	ProbeStatusStopped  = "stopped"
)

type ProbeStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
