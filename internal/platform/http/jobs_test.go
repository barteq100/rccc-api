package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/barteq100/rccc-api/internal/jobs"
)

func TestJobsHandlerListsJobsWithFiltersAndPagination(t *testing.T) {
	repo := jobs.NewMemoryRepository()
	seedHTTPJobs(t, repo)
	handler := NewJobsHandler(jobs.NewBrowseService(repo))

	req := httptest.NewRequest(http.MethodGet, "/v1/jobs?keyword=go&remote=true&source=greenhouse&seniority=senior&page=1&page_size=1", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d with body %s", res.Code, res.Body.String())
	}

	var response struct {
		Page     int              `json:"page"`
		PageSize int              `json:"page_size"`
		Total    int              `json:"total"`
		Items    []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if response.Total != 2 {
		t.Fatalf("expected total 2, got %d", response.Total)
	}
	if len(response.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(response.Items))
	}
	if response.Items[0]["id"] != "job-003" {
		t.Fatalf("expected latest matching job first, got %#v", response.Items[0])
	}
}

func TestJobsHandlerReturnsDetailByID(t *testing.T) {
	repo := jobs.NewMemoryRepository()
	seedHTTPJobs(t, repo)
	handler := NewJobsHandler(jobs.NewBrowseService(repo))

	req := httptest.NewRequest(http.MethodGet, "/v1/jobs/job-002", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d with body %s", res.Code, res.Body.String())
	}

	var response map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if response["id"] != "job-002" {
		t.Fatalf("expected job-002, got %#v", response)
	}
}

func TestJobsHandlerReturnsNotFoundForMissingJob(t *testing.T) {
	handler := NewJobsHandler(jobs.NewBrowseService(jobs.NewMemoryRepository()))

	req := httptest.NewRequest(http.MethodGet, "/v1/jobs/missing-job", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d with body %s", res.Code, res.Body.String())
	}
}

func seedHTTPJobs(t *testing.T, repo *jobs.MemoryRepository) {
	t.Helper()
	items := []jobs.Job{
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
	for _, item := range items {
		if err := repo.Upsert(reqContext(), item); err != nil {
			t.Fatalf("Upsert returned error: %v", err)
		}
	}
}

func reqContext() context.Context {
	return context.Background()
}


