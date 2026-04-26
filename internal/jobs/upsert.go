package jobs

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// UpsertJobInput is the canonical job payload accepted from ingestion.
type UpsertJobInput struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	Remote      bool      `json:"remote"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	SourceURL   string    `json:"source_url"`
	PostedAt    time.Time `json:"posted_at"`
	IngestedAt  time.Time `json:"ingested_at"`
}

// UpsertJobsResult summarizes a completed ingestion upsert batch.
type UpsertJobsResult struct {
	Received int `json:"received"`
	Upserted int `json:"upserted"`
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

// Repository is the persistence boundary for canonical jobs writes.
type Repository interface {
	Upsert(context.Context, Job) error
	GetByID(context.Context, string) (Job, bool, error)
}

// Clock provides the current time for persistence operations.
type Clock func() time.Time

// UpsertService validates canonical job payloads and persists them idempotently.
type UpsertService struct {
	repo  Repository
	clock Clock
}

// NewUpsertService constructs a jobs upsert service.
func NewUpsertService(repo Repository, clock Clock) *UpsertService {
	if clock == nil {
		clock = time.Now
	}

	return &UpsertService{repo: repo, clock: clock}
}

// Upsert validates and upserts a canonical jobs batch.
func (s *UpsertService) Upsert(ctx context.Context, inputs []UpsertJobInput) (UpsertJobsResult, error) {
	if len(inputs) == 0 {
		return UpsertJobsResult{}, ValidationError{Issues: []ValidationIssue{{
			Field:   "jobs",
			Message: "must contain at least one job",
		}}}
	}

	issues := make([]ValidationIssue, 0)
	for i, input := range inputs {
		issues = append(issues, validateInput(i, input)...)
	}

	if len(issues) > 0 {
		return UpsertJobsResult{}, ValidationError{Issues: issues}
	}

	now := s.clock().UTC()
	for _, input := range inputs {
		job := Job{
			ID:          strings.TrimSpace(input.ID),
			Title:       strings.TrimSpace(input.Title),
			Company:     strings.TrimSpace(input.Company),
			Location:    strings.TrimSpace(input.Location),
			Remote:      input.Remote,
			Description: strings.TrimSpace(input.Description),
			Source:      strings.TrimSpace(input.Source),
			SourceURL:   strings.TrimSpace(input.SourceURL),
			PostedAt:    input.PostedAt.UTC(),
			IngestedAt:  input.IngestedAt.UTC(),
			UpdatedAt:   now,
		}

		if err := s.repo.Upsert(ctx, job); err != nil {
			return UpsertJobsResult{}, err
		}
	}

	return UpsertJobsResult{Received: len(inputs), Upserted: len(inputs)}, nil
}

func validateInput(index int, input UpsertJobInput) []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	prefix := fmt.Sprintf("jobs[%d]", index)

	required := map[string]string{
		"id":          input.ID,
		"title":       input.Title,
		"company":     input.Company,
		"location":    input.Location,
		"description": input.Description,
		"source":      input.Source,
		"source_url":  input.SourceURL,
	}

	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			issues = append(issues, ValidationIssue{
				Field:   prefix + "." + field,
				Message: "must not be empty",
			})
		}
	}

	if input.PostedAt.IsZero() {
		issues = append(issues, ValidationIssue{Field: prefix + ".posted_at", Message: "must be a valid timestamp"})
	}

	if input.IngestedAt.IsZero() {
		issues = append(issues, ValidationIssue{Field: prefix + ".ingested_at", Message: "must be a valid timestamp"})
	}

	return issues
}

// MemoryRepository is a lightweight idempotent repository used by the current API scaffold and tests.
type MemoryRepository struct {
	mu   sync.RWMutex
	jobs map[string]Job
}

// NewMemoryRepository constructs an in-memory jobs repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{jobs: make(map[string]Job)}
}

// Upsert inserts or updates a canonical job by ID.
func (r *MemoryRepository) Upsert(_ context.Context, job Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := job.UpdatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}

	existing, found := r.jobs[job.ID]
	if found {
		job.CreatedAt = existing.CreatedAt
	} else {
		job.CreatedAt = now
	}
	job.UpdatedAt = now
	r.jobs[job.ID] = job
	return nil
}

// GetByID returns a previously persisted job.
func (r *MemoryRepository) GetByID(_ context.Context, id string) (Job, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	job, ok := r.jobs[id]
	return job, ok, nil
}

// List returns paginated jobs using the supplied server-side filters.
func (r *MemoryRepository) List(_ context.Context, query ListJobsQuery) ([]Job, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	filtered := make([]Job, 0, len(r.jobs))
	for _, job := range r.jobs {
		if !matchesListJobsQuery(job, query) {
			continue
		}
		filtered = append(filtered, job)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if !filtered[i].PostedAt.Equal(filtered[j].PostedAt) {
			return filtered[i].PostedAt.After(filtered[j].PostedAt)
		}
		if !filtered[i].IngestedAt.Equal(filtered[j].IngestedAt) {
			return filtered[i].IngestedAt.After(filtered[j].IngestedAt)
		}
		return filtered[i].ID < filtered[j].ID
	})

	total := len(filtered)
	start := (query.Page - 1) * query.PageSize
	if start >= total {
		return []Job{}, total, nil
	}
	end := start + query.PageSize
	if end > total {
		end = total
	}

	items := make([]Job, end-start)
	copy(items, filtered[start:end])
	return items, total, nil
}

func matchesListJobsQuery(job Job, query ListJobsQuery) bool {
	if query.Remote != nil && job.Remote != *query.Remote {
		return false
	}
	if query.Source != "" && !strings.EqualFold(strings.TrimSpace(job.Source), query.Source) {
		return false
	}

	if query.Keyword != "" {
		needle := strings.ToLower(query.Keyword)
		haystack := strings.ToLower(job.Title + "\n" + job.Company + "\n" + job.Description)
		if !strings.Contains(haystack, needle) {
			return false
		}
	}

	if query.SeniorityHint != "" {
		needle := strings.ToLower(query.SeniorityHint)
		haystack := strings.ToLower(job.Title + "\n" + job.Description)
		if !strings.Contains(haystack, needle) {
			return false
		}
	}

	return true
}
