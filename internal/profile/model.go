package profile

import "time"

// Preferences captures the single-user MVP profile configuration.
type Preferences struct {
	ID                 int16
	PreferredStack     []string
	RemoteOnly         bool
	PreferredLocations []string
	TargetSeniority    string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
