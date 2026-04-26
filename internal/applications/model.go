package applications

import "time"

// Status is the allowed application state for the MVP workflow.
type Status string

const (
	StatusSaved     Status = "saved"
	StatusApplied   Status = "applied"
	StatusInterview Status = "interview"
	StatusOffer     Status = "offer"
	StatusRejected  Status = "rejected"
)

// AllStatuses returns the fixed set of allowed application states.
func AllStatuses() []Status {
	return []Status{
		StatusSaved,
		StatusApplied,
		StatusInterview,
		StatusOffer,
		StatusRejected,
	}
}

// Valid reports whether the status is part of the fixed MVP state model.
func (s Status) Valid() bool {
	for _, allowed := range AllStatuses() {
		if s == allowed {
			return true
		}
	}

	return false
}

// Application is the persisted tracking record for a saved or applied job.
type Application struct {
	JobID           string
	Status          Status
	SavedAt         time.Time
	AppliedAt       *time.Time
	StatusChangedAt time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
