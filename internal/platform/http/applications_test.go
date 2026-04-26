package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/barteq100/rccc-api/internal/applications"
	"github.com/barteq100/rccc-api/internal/jobs"
)

func TestSavedJobsHandlerCreatesSavedJob(t *testing.T) {
	handler := NewSavedJobsHandler(newApplicationsServiceForHTTP(t))

	req := httptest.NewRequest(http.MethodPut, "/v1/saved-jobs/job-001", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d with body %s", res.Code, res.Body.String())
	}

	var response applicationResponse
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if response.Job.ID != "job-001" {
		t.Fatalf("expected job-001, got %#v", response)
	}
	if response.Status != "saved" {
		t.Fatalf("expected saved status, got %q", response.Status)
	}
	if response.AppliedAt != nil {
		t.Fatalf("expected applied_at to be omitted, got %v", *response.AppliedAt)
	}
}

func TestApplicationsHandlerMarksJobApplied(t *testing.T) {
	service := newApplicationsServiceForHTTP(t)
	if _, err := service.Save(context.Background(), "job-001"); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	handler := NewApplicationsHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/v1/applications/job-001/apply", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d with body %s", res.Code, res.Body.String())
	}

	var response applicationResponse
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if response.Status != "applied" {
		t.Fatalf("expected applied status, got %q", response.Status)
	}
	if response.AppliedAt == nil {
		t.Fatalf("expected applied_at to be set")
	}
}

func TestApplicationsHandlerUpdatesStatus(t *testing.T) {
	service := newApplicationsServiceForHTTP(t)
	handler := NewApplicationsHandler(service)

	payload := bytes.NewBufferString(`{"status":"interview"}`)
	req := httptest.NewRequest(http.MethodPatch, "/v1/applications/job-002", payload)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d with body %s", res.Code, res.Body.String())
	}

	var response applicationResponse
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if response.Job.ID != "job-002" || response.Status != "interview" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestApplicationsHandlerRejectsInvalidStatus(t *testing.T) {
	handler := NewApplicationsHandler(newApplicationsServiceForHTTP(t))

	req := httptest.NewRequest(http.MethodPatch, "/v1/applications/job-001", bytes.NewBufferString(`{"status":"draft"}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d with body %s", res.Code, res.Body.String())
	}

	var response map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if response["error"] != "validation_failed" {
		t.Fatalf("expected validation_failed error, got %#v", response)
	}
}

func TestApplicationsHandlerListsTrackedApplications(t *testing.T) {
	service := newApplicationsServiceForHTTP(t)
	if _, err := service.Save(context.Background(), "job-001"); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if _, err := service.UpdateStatus(context.Background(), "job-002", applications.StatusInterview); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}
	handler := NewApplicationsHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/v1/applications", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d with body %s", res.Code, res.Body.String())
	}

	var response listApplicationsResponse
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if len(response.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(response.Items))
	}
	if response.Items[0].Job.ID != "job-002" || response.Items[0].Status != "interview" {
		t.Fatalf("expected most recent status first, got %#v", response.Items[0])
	}
}

func newApplicationsServiceForHTTP(t *testing.T) *applications.Service {
	t.Helper()

	jobRepo := jobs.NewMemoryRepository()
	seedApplicationHTTPJob(t, jobRepo, "job-001", time.Date(2026, 4, 24, 9, 0, 0, 0, time.UTC))
	seedApplicationHTTPJob(t, jobRepo, "job-002", time.Date(2026, 4, 24, 9, 5, 0, 0, time.UTC))

	return applications.NewService(applications.NewMemoryRepository(), jobRepo, fixedHTTPClock(
		time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 24, 12, 5, 0, 0, time.UTC),
		time.Date(2026, 4, 24, 12, 10, 0, 0, time.UTC),
	))
}

func seedApplicationHTTPJob(t *testing.T, repo *jobs.MemoryRepository, id string, postedAt time.Time) {
	t.Helper()

	if err := repo.Upsert(context.Background(), jobs.Job{
		ID:          id,
		Title:       "Backend Engineer " + id,
		Company:     "Acme",
		Location:    "Remote - Europe",
		Remote:      true,
		Description: "Build backend services in Go.",
		Source:      "greenhouse",
		SourceURL:   "https://example.com/" + id,
		PostedAt:    postedAt,
		IngestedAt:  postedAt,
		UpdatedAt:   postedAt,
	}); err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}
}

func fixedHTTPClock(values ...time.Time) func() time.Time {
	index := 0

	return func() time.Time {
		if index >= len(values) {
			return values[len(values)-1]
		}
		value := values[index]
		index++
		return value
	}
}
