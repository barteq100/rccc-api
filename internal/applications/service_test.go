package applications

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/barteq100/rccc-api/internal/jobs"
)

func TestServiceSaveCreatesSavedApplication(t *testing.T) {
	jobRepo := jobs.NewMemoryRepository()
	seedApplicationTestJob(t, jobRepo, "job-001")

	clock := fixedClock(
		time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC),
	)
	service := NewService(NewMemoryRepository(), jobRepo, clock)

	got, err := service.Save(context.Background(), " job-001 ")
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	if got.Status != StatusSaved {
		t.Fatalf("expected saved status, got %q", got.Status)
	}
	if got.AppliedAt != nil {
		t.Fatalf("expected applied_at to be nil, got %v", got.AppliedAt)
	}
	if got.SavedAt != time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC) {
		t.Fatalf("unexpected saved_at: %s", got.SavedAt)
	}
}

func TestServiceMarkAppliedPromotesSavedJob(t *testing.T) {
	jobRepo := jobs.NewMemoryRepository()
	seedApplicationTestJob(t, jobRepo, "job-001")

	clock := fixedClock(
		time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 24, 12, 5, 0, 0, time.UTC),
	)
	service := NewService(NewMemoryRepository(), jobRepo, clock)

	if _, err := service.Save(context.Background(), "job-001"); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	got, err := service.MarkApplied(context.Background(), "job-001")
	if err != nil {
		t.Fatalf("MarkApplied returned error: %v", err)
	}

	if got.Status != StatusApplied {
		t.Fatalf("expected applied status, got %q", got.Status)
	}
	if got.AppliedAt == nil {
		t.Fatalf("expected applied_at to be set")
	}
	if got.SavedAt != time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC) {
		t.Fatalf("unexpected saved_at: %s", got.SavedAt)
	}
	if got.AppliedAt.UTC() != time.Date(2026, 4, 24, 12, 5, 0, 0, time.UTC) {
		t.Fatalf("unexpected applied_at: %s", got.AppliedAt.UTC())
	}
}

func TestServiceUpdateStatusRejectsInvalidStatus(t *testing.T) {
	jobRepo := jobs.NewMemoryRepository()
	seedApplicationTestJob(t, jobRepo, "job-001")

	service := NewService(NewMemoryRepository(), jobRepo, fixedClock(time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)))

	_, err := service.UpdateStatus(context.Background(), "job-001", Status("draft"))
	if err == nil {
		t.Fatal("expected validation error")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if len(validationErr.Issues) != 1 || validationErr.Issues[0].Field != "status" {
		t.Fatalf("unexpected validation issues: %#v", validationErr.Issues)
	}
}

func TestServiceListReturnsTrackedApplicationsSortedByStatusChange(t *testing.T) {
	jobRepo := jobs.NewMemoryRepository()
	seedApplicationTestJob(t, jobRepo, "job-001")
	seedApplicationTestJob(t, jobRepo, "job-002")

	clock := fixedClock(
		time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 24, 12, 10, 0, 0, time.UTC),
	)
	service := NewService(NewMemoryRepository(), jobRepo, clock)

	if _, err := service.Save(context.Background(), "job-001"); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if _, err := service.UpdateStatus(context.Background(), "job-002", StatusInterview); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}

	items, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 tracked applications, got %d", len(items))
	}
	if items[0].Job.ID != "job-002" || items[0].Status != StatusInterview {
		t.Fatalf("expected most recent status change first, got %#v", items[0])
	}
	if items[1].Job.ID != "job-001" || items[1].Status != StatusSaved {
		t.Fatalf("unexpected second item: %#v", items[1])
	}
}

func TestServiceReturnsJobNotFoundForUnknownJob(t *testing.T) {
	service := NewService(NewMemoryRepository(), jobs.NewMemoryRepository(), fixedClock(time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)))

	_, err := service.Save(context.Background(), "missing-job")
	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got %v", err)
	}
}

func seedApplicationTestJob(t *testing.T, repo *jobs.MemoryRepository, id string) {
	t.Helper()

	now := time.Date(2026, 4, 24, 9, 0, 0, 0, time.UTC)
	if err := repo.Upsert(context.Background(), jobs.Job{
		ID:          id,
		Title:       "Backend Engineer " + id,
		Company:     "Acme",
		Location:    "Remote - Europe",
		Remote:      true,
		Description: "Build backend services in Go.",
		Source:      "greenhouse",
		SourceURL:   "https://example.com/" + id,
		PostedAt:    now,
		IngestedAt:  now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}
}

func fixedClock(values ...time.Time) func() time.Time {
	index := 0

	return func() time.Time {
		if len(values) == 0 {
			return time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
		}
		if index >= len(values) {
			return values[len(values)-1]
		}

		value := values[index]
		index++
		return value
	}
}
