package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/barteq100/rccc-api/internal/jobs"
)

func TestJobsUpsertHandlerReturnsSuccessForCanonicalPayload(t *testing.T) {
	repo := jobs.NewMemoryRepository()
	handler := NewJobsUpsertHandler(jobs.NewUpsertService(repo, func() time.Time {
		return time.Date(2026, 3, 23, 15, 0, 0, 0, time.UTC)
	}))

	body := map[string]any{
		"jobs": []map[string]any{{
			"id":          "gh-acme-senior-go-001",
			"title":       "Senior Go Engineer",
			"company":     "Acme",
			"location":    "Remote - Poland",
			"remote":      true,
			"description": "Build backend services for remote teams.",
			"source":      "greenhouse",
			"source_url":  "https://boards.greenhouse.io/acme/jobs/1",
			"posted_at":   "2026-03-22T10:00:00Z",
			"ingested_at": "2026-03-23T08:00:00Z",
		}},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/internal/ingestion/jobs", bytes.NewReader(payload))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d with body %s", res.Code, res.Body.String())
	}

	var response jobs.UpsertJobsResult
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if response.Received != 1 || response.Upserted != 1 {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestJobsUpsertHandlerReturnsValidationErrors(t *testing.T) {
	handler := NewJobsUpsertHandler(jobs.NewUpsertService(jobs.NewMemoryRepository(), time.Now))

	req := httptest.NewRequest(http.MethodPost, "/internal/ingestion/jobs", bytes.NewBufferString(`{"jobs":[{"id":"","title":""}]}`))
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
