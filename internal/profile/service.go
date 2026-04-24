package profile

import (
	"context"
	"strings"
	"sync"
	"time"
)

const singletonProfileID int16 = 1

// UpdateInput is the user-facing profile preferences payload.
type UpdateInput struct {
	PreferredStack     []string
	RemoteOnly         bool
	PreferredLocations []string
	TargetSeniority    string
}

// ValidationIssue describes a single request validation failure.
type ValidationIssue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError returns structured validation issues to the HTTP boundary.
type ValidationError struct {
	Issues []ValidationIssue `json:"issues"`
}

func (e ValidationError) Error() string {
	return "request validation failed"
}

// Repository is the persistence boundary for the single-user profile.
type Repository interface {
	Get(context.Context) (Preferences, error)
	Save(context.Context, Preferences) (Preferences, error)
}

// Clock provides the current time for persistence operations.
type Clock func() time.Time

// Service provides single-user profile preference reads and updates.
type Service struct {
	repo  Repository
	clock Clock
}

// NewService constructs a profile service.
func NewService(repo Repository, clock Clock) *Service {
	if clock == nil {
		clock = time.Now
	}

	return &Service{repo: repo, clock: clock}
}

// Get returns the current single-user preferences.
func (s *Service) Get(ctx context.Context) (Preferences, error) {
	return s.repo.Get(ctx)
}

// Update normalizes and persists the single-user preferences.
func (s *Service) Update(ctx context.Context, input UpdateInput) (Preferences, error) {
	normalized, issues := normalizeUpdateInput(input)
	if len(issues) > 0 {
		return Preferences{}, ValidationError{Issues: issues}
	}

	current, err := s.repo.Get(ctx)
	if err != nil {
		return Preferences{}, err
	}

	now := s.clock().UTC()
	current.ID = singletonProfileID
	current.PreferredStack = normalized.PreferredStack
	current.RemoteOnly = normalized.RemoteOnly
	current.PreferredLocations = normalized.PreferredLocations
	current.TargetSeniority = normalized.TargetSeniority
	if current.CreatedAt.IsZero() {
		current.CreatedAt = now
	}
	current.UpdatedAt = now

	return s.repo.Save(ctx, current)
}

func normalizeUpdateInput(input UpdateInput) (UpdateInput, []ValidationIssue) {
	normalized := UpdateInput{
		PreferredStack:     normalizeStringList(input.PreferredStack),
		RemoteOnly:         input.RemoteOnly,
		PreferredLocations: normalizeStringList(input.PreferredLocations),
		TargetSeniority:    strings.TrimSpace(input.TargetSeniority),
	}

	return normalized, nil
}

func normalizeStringList(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	return normalized
}

// MemoryRepository is a lightweight single-user profile store used by the current API scaffold and tests.
type MemoryRepository struct {
	mu          sync.RWMutex
	preferences Preferences
	initialized bool
}

// NewMemoryRepository constructs an in-memory profile repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

// Get returns the single-user preferences, defaulting to an empty profile.
func (r *MemoryRepository) Get(_ context.Context) (Preferences, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return Preferences{
			ID:                 singletonProfileID,
			PreferredStack:     []string{},
			PreferredLocations: []string{},
		}, nil
	}

	return clonePreferences(r.preferences), nil
}

// Save persists the single-user preferences.
func (r *MemoryRepository) Save(_ context.Context, preferences Preferences) (Preferences, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if preferences.ID == 0 {
		preferences.ID = singletonProfileID
	}
	if preferences.PreferredStack == nil {
		preferences.PreferredStack = []string{}
	}
	if preferences.PreferredLocations == nil {
		preferences.PreferredLocations = []string{}
	}

	r.preferences = clonePreferences(preferences)
	r.initialized = true
	return clonePreferences(r.preferences), nil
}

func clonePreferences(preferences Preferences) Preferences {
	preferences.PreferredStack = append([]string{}, preferences.PreferredStack...)
	preferences.PreferredLocations = append([]string{}, preferences.PreferredLocations...)
	return preferences
}
