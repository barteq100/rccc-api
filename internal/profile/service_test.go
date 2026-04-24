package profile

import (
	"context"
	"testing"
	"time"
)

func TestServiceGetReturnsDefaultPreferences(t *testing.T) {
	service := NewService(NewMemoryRepository(), time.Now)

	result, err := service.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if result.ID != 1 {
		t.Fatalf("expected singleton profile id 1, got %d", result.ID)
	}
	if result.RemoteOnly {
		t.Fatalf("expected remote_only to default to false")
	}
	if len(result.PreferredStack) != 0 {
		t.Fatalf("expected empty preferred stack, got %#v", result.PreferredStack)
	}
	if len(result.PreferredLocations) != 0 {
		t.Fatalf("expected empty preferred locations, got %#v", result.PreferredLocations)
	}
	if result.TargetSeniority != "" {
		t.Fatalf("expected empty target seniority, got %q", result.TargetSeniority)
	}
}

func TestServiceUpdateNormalizesAndPersistsPreferences(t *testing.T) {
	repo := NewMemoryRepository()
	service := NewService(repo, func() time.Time {
		return time.Date(2026, 4, 24, 11, 30, 0, 0, time.UTC)
	})

	result, err := service.Update(context.Background(), UpdateInput{
		PreferredStack:     []string{" Go ", "TypeScript", "go", ""},
		RemoteOnly:         true,
		PreferredLocations: []string{" Remote - Europe ", "Warsaw", "warsaw"},
		TargetSeniority:    "  senior ",
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	expectedStack := []string{"Go", "TypeScript"}
	if len(result.PreferredStack) != len(expectedStack) {
		t.Fatalf("unexpected preferred stack: %#v", result.PreferredStack)
	}
	for i := range expectedStack {
		if result.PreferredStack[i] != expectedStack[i] {
			t.Fatalf("unexpected preferred stack at %d: got %q want %q", i, result.PreferredStack[i], expectedStack[i])
		}
	}

	expectedLocations := []string{"Remote - Europe", "Warsaw"}
	if len(result.PreferredLocations) != len(expectedLocations) {
		t.Fatalf("unexpected preferred locations: %#v", result.PreferredLocations)
	}
	for i := range expectedLocations {
		if result.PreferredLocations[i] != expectedLocations[i] {
			t.Fatalf("unexpected preferred locations at %d: got %q want %q", i, result.PreferredLocations[i], expectedLocations[i])
		}
	}

	if !result.RemoteOnly {
		t.Fatalf("expected remote_only to be true")
	}
	if result.TargetSeniority != "senior" {
		t.Fatalf("unexpected target seniority: %q", result.TargetSeniority)
	}
	if result.CreatedAt.IsZero() {
		t.Fatalf("expected created_at to be set")
	}
	if !result.UpdatedAt.Equal(time.Date(2026, 4, 24, 11, 30, 0, 0, time.UTC)) {
		t.Fatalf("unexpected updated_at: %s", result.UpdatedAt)
	}

	stored, err := repo.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if stored.TargetSeniority != "senior" {
		t.Fatalf("expected stored target seniority to persist, got %q", stored.TargetSeniority)
	}
}
