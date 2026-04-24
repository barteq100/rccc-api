package jobs

import (
	"context"
	"strings"

	"github.com/barteq100/rccc-api/internal/profile"
	"github.com/barteq100/rccc-api/internal/scoring"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// ListJobsQuery defines the user-facing browse filters and pagination.
type ListJobsQuery struct {
	Keyword       string
	Remote        *bool
	Source        string
	SeniorityHint string
	Page          int
	PageSize      int
}

// ListJobsResult is the paginated response payload returned by the browse service.
type ListJobsResult struct {
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Total    int         `json:"total"`
	Items    []JobResult `json:"items"`
}

// JobResult is the user-facing scored job payload.
type JobResult struct {
	Job          Job      `json:"job"`
	Score        int      `json:"score"`
	ScoreReasons []string `json:"score_reasons"`
}

// PreferencesProvider exposes the single-user preferences needed for score calculation.
type PreferencesProvider interface {
	Get(context.Context) (profile.Preferences, error)
}

// FitScorer evaluates a deterministic fit score for a canonical job.
type FitScorer interface {
	Evaluate(scoring.JobInput, scoring.Preferences) scoring.Result
}

// QueryRepository exposes the read-side jobs queries needed by the API.
type QueryRepository interface {
	GetByID(context.Context, string) (Job, bool, error)
	List(context.Context, ListJobsQuery) ([]Job, int, error)
}

// BrowseService provides user-facing browse and detail behavior for jobs.
type BrowseService struct {
	repo        QueryRepository
	preferences PreferencesProvider
	scorer      FitScorer
}

// NewBrowseService constructs a jobs browse service.
func NewBrowseService(repo QueryRepository, preferences PreferencesProvider, scorer FitScorer) *BrowseService {
	if scorer == nil {
		scorer = scoring.NewService()
	}

	return &BrowseService{
		repo:        repo,
		preferences: preferences,
		scorer:      scorer,
	}
}

// List returns paginated jobs for the user-facing browse endpoint.
func (s *BrowseService) List(ctx context.Context, query ListJobsQuery) (ListJobsResult, error) {
	normalized := normalizeListJobsQuery(query)
	items, total, err := s.repo.List(ctx, normalized)
	if err != nil {
		return ListJobsResult{}, err
	}

	scoredItems, err := s.scoreJobs(ctx, items)
	if err != nil {
		return ListJobsResult{}, err
	}

	return ListJobsResult{
		Page:     normalized.Page,
		PageSize: normalized.PageSize,
		Total:    total,
		Items:    scoredItems,
	}, nil
}

// GetByID returns a single job detail by id.
func (s *BrowseService) GetByID(ctx context.Context, id string) (JobResult, bool, error) {
	job, found, err := s.repo.GetByID(ctx, strings.TrimSpace(id))
	if err != nil || !found {
		return JobResult{}, found, err
	}

	scoredItems, err := s.scoreJobs(ctx, []Job{job})
	if err != nil {
		return JobResult{}, false, err
	}

	return scoredItems[0], true, nil
}

func normalizeListJobsQuery(query ListJobsQuery) ListJobsQuery {
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.Source = strings.TrimSpace(query.Source)
	query.SeniorityHint = strings.TrimSpace(query.SeniorityHint)
	if query.Page <= 0 {
		query.Page = defaultPage
	}
	if query.PageSize <= 0 {
		query.PageSize = defaultPageSize
	}
	if query.PageSize > maxPageSize {
		query.PageSize = maxPageSize
	}
	return query
}

func (s *BrowseService) scoreJobs(ctx context.Context, jobsList []Job) ([]JobResult, error) {
	preferences, err := s.loadPreferences(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]JobResult, 0, len(jobsList))
	for _, job := range jobsList {
		score := s.scorer.Evaluate(scoring.JobInput{
			Title:       job.Title,
			Location:    job.Location,
			Remote:      job.Remote,
			Description: job.Description,
		}, scoring.Preferences{
			PreferredStack:     preferences.PreferredStack,
			RemoteOnly:         preferences.RemoteOnly,
			PreferredLocations: preferences.PreferredLocations,
			TargetSeniority:    preferences.TargetSeniority,
		})

		results = append(results, JobResult{
			Job:          job,
			Score:        score.Score,
			ScoreReasons: append([]string{}, score.Reasons...),
		})
	}

	return results, nil
}

func (s *BrowseService) loadPreferences(ctx context.Context) (profile.Preferences, error) {
	if s.preferences == nil {
		return profile.Preferences{
			PreferredStack:     []string{},
			PreferredLocations: []string{},
		}, nil
	}

	preferences, err := s.preferences.Get(ctx)
	if err != nil {
		return profile.Preferences{}, err
	}

	if preferences.PreferredStack == nil {
		preferences.PreferredStack = []string{}
	}
	if preferences.PreferredLocations == nil {
		preferences.PreferredLocations = []string{}
	}

	return preferences, nil
}
