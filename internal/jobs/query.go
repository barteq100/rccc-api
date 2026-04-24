package jobs

import (
	"context"
	"strings"
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
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int   `json:"total"`
	Items    []Job `json:"items"`
}

// QueryRepository exposes the read-side jobs queries needed by the API.
type QueryRepository interface {
	GetByID(context.Context, string) (Job, bool, error)
	List(context.Context, ListJobsQuery) ([]Job, int, error)
}

// BrowseService provides user-facing browse and detail behavior for jobs.
type BrowseService struct {
	repo QueryRepository
}

// NewBrowseService constructs a jobs browse service.
func NewBrowseService(repo QueryRepository) *BrowseService {
	return &BrowseService{repo: repo}
}

// List returns paginated jobs for the user-facing browse endpoint.
func (s *BrowseService) List(ctx context.Context, query ListJobsQuery) (ListJobsResult, error) {
	normalized := normalizeListJobsQuery(query)
	items, total, err := s.repo.List(ctx, normalized)
	if err != nil {
		return ListJobsResult{}, err
	}

	return ListJobsResult{
		Page:     normalized.Page,
		PageSize: normalized.PageSize,
		Total:    total,
		Items:    items,
	}, nil
}

// GetByID returns a single job detail by id.
func (s *BrowseService) GetByID(ctx context.Context, id string) (Job, bool, error) {
	return s.repo.GetByID(ctx, strings.TrimSpace(id))
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
