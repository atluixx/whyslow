package models

// Diagnosis is an actionable observation about a possible bottleneck.
type Diagnosis struct {
	Severity     Severity
	Title        string
	Message      string
	Contributors []ProcessStats
}
