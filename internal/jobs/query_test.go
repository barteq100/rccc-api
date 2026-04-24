package jobs

import (
	"context"
	"testing"
	"time"
)

func TestBrowseServiceListAppliesFiltersAndPagination(t *testing.T) {
	repo := NewMemoryRepository()
	seedJobs(t, repo)
	service := NewBrowseService(repo)
	remote := true

	result, err := service.List(context.Background(), ListJobsQuery{
		Keyword:       "go",
		Remote:        &remote,
		Source:        "greenhouse",
		SeniorityHint: "senior",
		Page:          1,
		PageSize:      1,
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if result.Total != 2 {
		t.Fatalf("expected total 2, got %d", result.Total)
	}
	if result.Page != 1 || result.PageSize != 1 {
		t.Fatalf("unexpected pagination result: %#v", result)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected one item on first page, got %d", len(result.Items))
	}
	if result.Items[0].ID != "job-003" {
		t.Fatalf("expected most recent matching job first, got %q", result.Items[0].ID)
	}
}

func TestBrowseServiceGetByIDReturnsJob(t *testing.T) {
	repo := NewMemoryRepository()
	seedJobs(t, repo)
	service := NewBrowseService(repo)

	job, found, err := service.GetByID(context.Background(), "job-002")
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if !found {
		t.Fatalf("expected job to be found")
	}
	if job.Title != "Staff Platform Engineer" {
		t.Fatalf("unexpected job title: %q", job.Title)
	}
}

func seedJobs(t *testing.T, repo *MemoryRepository) {
	t.Helper()
	jobs := []Job{
		{
			ID:          "job-001",
			Title:       "Senior Go Engineer",
			Company:     "Acme",
			Location:    "Remote - Poland",
			Remote:      true,
			Description: "Build backend services in Go.",
			Source:      "greenhouse",
			SourceURL:   "https://example.com/1",
			PostedAt:    time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
			IngestedAt:  time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:          "job-002",
			Title:       "Staff Platform Engineer",
			Company:     "Beta",
			Location:    "Warsaw",
			Remote:      false,
			Description: "Platform engineering and internal tooling.",
			Source:      "lever",
			SourceURL:   "https://example.com/2",
			PostedAt:    time.Date(2026, 3, 19, 10, 0, 0, 0, time.UTC),
			IngestedAt:  time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:          "job-003",
			Title:       "Senior Go Platform Engineer",
			Company:     "Acme",
			Location:    "Remote - Europe",
			Remote:      true,
			Description: "Senior Go and platform work for distributed systems.",
			Source:      "greenhouse",
			SourceURL:   "https://example.com/3",
			PostedAt:    time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC),
			IngestedAt:  time.Date(2026, 3, 23, 10, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 3, 23, 10, 0, 0, 0, time.UTC),
		},
	}

	for _, job := range jobs {
		if err := repo.Upsert(context.Background(), job); err != nil {
			t.Fatalf("Upsert returned error: %v", err)
		}
	}
}
