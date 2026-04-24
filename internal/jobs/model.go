package jobs

import "time"

// Job is the canonical job record stored by rccc-api.
type Job struct {
	ID          string
	Title       string
	Company     string
	Location    string
	Remote      bool
	Description string
	Source      string
	SourceURL   string
	PostedAt    time.Time
	IngestedAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
