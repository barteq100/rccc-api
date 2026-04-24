package applications

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/barteq100/rccc-api/internal/jobs"
)

var ErrJobNotFound = errors.New("job not found")

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

// Repository is the persistence boundary for saved jobs and application tracking.
type Repository interface {
	Upsert(context.Context, Application) error
	GetByJobID(context.Context, string) (Application, bool, error)
	List(context.Context) ([]Application, error)
}

// JobRepository exposes the canonical job reads needed by application tracking.
type JobRepository interface {
	GetByID(context.Context, string) (jobs.Job, bool, error)
}

// Clock provides the current time for application tracking operations.
type Clock func() time.Time

// TrackedApplication is the user-facing application record enriched with job data.
type TrackedApplication struct {
	Job             jobs.Job
	Status          Status
	SavedAt         time.Time
	AppliedAt       *time.Time
	StatusChangedAt time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Service provides save, apply, status-update, and tracked-list behavior.
type Service struct {
	repo    Repository
	jobRepo JobRepository
	clock   Clock
}

// NewService constructs an applications service.
func NewService(repo Repository, jobRepo JobRepository, clock Clock) *Service {
	if clock == nil {
		clock = time.Now
	}

	return &Service{
		repo:    repo,
		jobRepo: jobRepo,
		clock:   clock,
	}
}

// Save records a saved job without overwriting an existing tracked status.
func (s *Service) Save(ctx context.Context, jobID string) (TrackedApplication, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return TrackedApplication{}, ValidationError{Issues: []ValidationIssue{{
			Field:   "job_id",
			Message: "must not be empty",
		}}}
	}

	job, err := s.requireJob(ctx, jobID)
	if err != nil {
		return TrackedApplication{}, err
	}

	existing, found, err := s.repo.GetByJobID(ctx, jobID)
	if err != nil {
		return TrackedApplication{}, err
	}
	if found {
		return track(job, existing), nil
	}

	now := s.clock().UTC()
	record := Application{
		JobID:           jobID,
		Status:          StatusSaved,
		SavedAt:         now,
		StatusChangedAt: now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.Upsert(ctx, record); err != nil {
		return TrackedApplication{}, err
	}

	return track(job, record), nil
}

// MarkApplied creates or updates a tracked job to the applied status.
func (s *Service) MarkApplied(ctx context.Context, jobID string) (TrackedApplication, error) {
	return s.UpdateStatus(ctx, jobID, StatusApplied)
}

// UpdateStatus sets the tracked status using the fixed MVP state model.
func (s *Service) UpdateStatus(ctx context.Context, jobID string, status Status) (TrackedApplication, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return TrackedApplication{}, ValidationError{Issues: []ValidationIssue{{
			Field:   "job_id",
			Message: "must not be empty",
		}}}
	}
	if !status.Valid() {
		return TrackedApplication{}, ValidationError{Issues: []ValidationIssue{{
			Field:   "status",
			Message: "must be one of saved, applied, interview, offer, rejected",
		}}}
	}

	job, err := s.requireJob(ctx, jobID)
	if err != nil {
		return TrackedApplication{}, err
	}

	existing, found, err := s.repo.GetByJobID(ctx, jobID)
	if err != nil {
		return TrackedApplication{}, err
	}

	now := s.clock().UTC()
	record := Application{
		JobID:           jobID,
		Status:          status,
		StatusChangedAt: now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if found {
		record.SavedAt = existing.SavedAt
		record.CreatedAt = existing.CreatedAt
		if existing.Status == status && appliedAtMatchesStatus(existing.AppliedAt, status) {
			return track(job, existing), nil
		}
	} else {
		record.SavedAt = now
	}

	record.AppliedAt = appliedAtForStatus(status, existing.AppliedAt, now)
	if status == StatusSaved {
		record.AppliedAt = nil
	}

	if err := s.repo.Upsert(ctx, record); err != nil {
		return TrackedApplication{}, err
	}

	return track(job, record), nil
}

// List returns tracked applications enriched with their corresponding jobs.
func (s *Service) List(ctx context.Context) ([]TrackedApplication, error) {
	records, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]TrackedApplication, 0, len(records))
	for _, record := range records {
		job, err := s.requireJob(ctx, record.JobID)
		if err != nil {
			return nil, err
		}
		items = append(items, track(job, record))
	}

	sort.Slice(items, func(i, j int) bool {
		if !items[i].StatusChangedAt.Equal(items[j].StatusChangedAt) {
			return items[i].StatusChangedAt.After(items[j].StatusChangedAt)
		}
		if !items[i].SavedAt.Equal(items[j].SavedAt) {
			return items[i].SavedAt.After(items[j].SavedAt)
		}
		return items[i].Job.ID < items[j].Job.ID
	})

	return items, nil
}

func (s *Service) requireJob(ctx context.Context, jobID string) (jobs.Job, error) {
	if s.jobRepo == nil {
		return jobs.Job{}, errors.New("job repository is not configured")
	}

	job, found, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return jobs.Job{}, err
	}
	if !found {
		return jobs.Job{}, ErrJobNotFound
	}

	return job, nil
}

func track(job jobs.Job, application Application) TrackedApplication {
	return TrackedApplication{
		Job:             job,
		Status:          application.Status,
		SavedAt:         application.SavedAt,
		AppliedAt:       application.AppliedAt,
		StatusChangedAt: application.StatusChangedAt,
		CreatedAt:       application.CreatedAt,
		UpdatedAt:       application.UpdatedAt,
	}
}

func appliedAtForStatus(status Status, existing *time.Time, now time.Time) *time.Time {
	if status == StatusSaved {
		return nil
	}
	if existing != nil {
		value := existing.UTC()
		return &value
	}

	appliedAt := now
	return &appliedAt
}

func appliedAtMatchesStatus(appliedAt *time.Time, status Status) bool {
	if status == StatusSaved {
		return appliedAt == nil
	}

	return appliedAt != nil
}

// MemoryRepository is a lightweight repository used by the current API scaffold and tests.
type MemoryRepository struct {
	mu           sync.RWMutex
	applications map[string]Application
}

// NewMemoryRepository constructs an in-memory applications repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{applications: make(map[string]Application)}
}

// Upsert inserts or updates a tracked application by job ID.
func (r *MemoryRepository) Upsert(_ context.Context, application Application) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	record := application
	now := record.UpdatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if record.StatusChangedAt.IsZero() {
		record.StatusChangedAt = now
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	if record.SavedAt.IsZero() {
		record.SavedAt = now
	}

	if existing, found := r.applications[record.JobID]; found {
		record.CreatedAt = existing.CreatedAt
	}

	record.UpdatedAt = now
	r.applications[record.JobID] = record
	return nil
}

// GetByJobID returns a tracked application for a specific job.
func (r *MemoryRepository) GetByJobID(_ context.Context, jobID string) (Application, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, ok := r.applications[jobID]
	return record, ok, nil
}

// List returns all tracked applications.
func (r *MemoryRepository) List(_ context.Context) ([]Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]Application, 0, len(r.applications))
	for _, application := range r.applications {
		items = append(items, application)
	}

	return items, nil
}
