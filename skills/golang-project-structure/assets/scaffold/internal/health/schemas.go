package health

// Status values reported by the health check.
const (
	StatusOK       = "ok"
	StatusDegraded = "degraded"
)

// Report is the health check result.
type Report struct {
	Status  string            `json:"status"`
	Version string            `json:"version"`
	Checks  map[string]string `json:"checks"`
}
