package jobs

import (
	"context"
	"testing"
	"time"
)

func TestUpsertServicePersistsAndUpdatesExistingJob(t *testing.T) {
	repo := NewMemoryRepository()
	firstNow := time.Date(2026, 3, 23, 15, 0, 0, 0, time.UTC)
	service := NewUpsertService(repo, func() time.Time { return firstNow })

	result, err := service.Upsert(context.Background(), []UpsertJobInput{sampleInput()})
	if err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}

	if result.Received != 1 || result.Upserted != 1 {
		t.Fatalf("unexpected upsert result: %#v", result)
	}

	stored, found, err := repo.GetByID(context.Background(), "gh-acme-senior-go-001")
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if !found {
		t.Fatalf("expected stored job to exist")
	}
	if stored.Title != "Senior Go Engineer" {
		t.Fatalf("unexpected title after first upsert: %q", stored.Title)
	}
	if !stored.CreatedAt.Equal(firstNow) || !stored.UpdatedAt.Equal(firstNow) {
		t.Fatalf("unexpected timestamps after first upsert: %#v", stored)
	}

	secondNow := firstNow.Add(2 * time.Hour)
	service = NewUpsertService(repo, func() time.Time { return secondNow })
	updated := sampleInput()
	updated.Title = "Principal Go Engineer"

	if _, err := service.Upsert(context.Background(), []UpsertJobInput{updated}); err != nil {
		t.Fatalf("second Upsert returned error: %v", err)
	}

	stored, found, err = repo.GetByID(context.Background(), "gh-acme-senior-go-001")
	if err != nil {
		t.Fatalf("GetByID after update returned error: %v", err)
	}
	if !found {
		t.Fatalf("expected stored job to exist after update")
	}
	if stored.Title != "Principal Go Engineer" {
		t.Fatalf("expected upsert to update title, got %q", stored.Title)
	}
	if !stored.CreatedAt.Equal(firstNow) {
		t.Fatalf("expected created_at to remain stable, got %s", stored.CreatedAt)
	}
	if !stored.UpdatedAt.Equal(secondNow) {
		t.Fatalf("expected updated_at to refresh, got %s", stored.UpdatedAt)
	}
}

func TestUpsertServiceRejectsInvalidPayload(t *testing.T) {
	service := NewUpsertService(NewMemoryRepository(), func() time.Time { return time.Date(2026, 3, 23, 15, 0, 0, 0, time.UTC) })

	_, err := service.Upsert(context.Background(), []UpsertJobInput{{}})
	if err == nil {
		t.Fatalf("expected validation error")
	}

	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if len(validationErr.Issues) < 3 {
		t.Fatalf("expected multiple validation issues, got %#v", validationErr.Issues)
	}
}

func sampleInput() UpsertJobInput {
	return UpsertJobInput{
		ID:          "gh-acme-senior-go-001",
		Title:       "Senior Go Engineer",
		Company:     "Acme",
		Location:    "Remote - Poland",
		Remote:      true,
		Description: "Build backend services for remote teams.",
		Source:      "greenhouse",
		SourceURL:   "https://boards.greenhouse.io/acme/jobs/1",
		PostedAt:    time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC),
		IngestedAt:  time.Date(2026, 3, 23, 8, 0, 0, 0, time.UTC),
	}
}
