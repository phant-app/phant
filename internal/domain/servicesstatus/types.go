package servicesstatus

type State string

const (
	StateRunning     State = "running"
	StateStopped     State = "stopped"
	StateUnavailable State = "unavailable"
)

type ServiceStatus struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Port        int    `json:"port"`
	State       State  `json:"state"`
	Unit        string `json:"unit"`
}

type Snapshot struct {
	GeneratedAt string          `json:"generatedAt"`
	Platform    string          `json:"platform"`
	Supported   bool            `json:"supported"`
	Services    []ServiceStatus `json:"services"`
	Warnings    []string        `json:"warnings"`
	LastError   string          `json:"lastError"`
}
