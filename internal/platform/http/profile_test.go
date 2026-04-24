package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/barteq100/rccc-api/internal/profile"
)

func TestProfileHandlerReturnsDefaultPreferences(t *testing.T) {
	handler := NewProfileHandler(profile.NewService(profile.NewMemoryRepository(), time.Now))

	req := httptest.NewRequest(http.MethodGet, "/v1/profile/preferences", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d with body %s", res.Code, res.Body.String())
	}

	var response profilePreferencesResponse
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if response.RemoteOnly {
		t.Fatalf("expected remote_only to default to false")
	}
	if len(response.PreferredStack) != 0 {
		t.Fatalf("expected empty preferred stack, got %#v", response.PreferredStack)
	}
	if len(response.PreferredLocations) != 0 {
		t.Fatalf("expected empty preferred locations, got %#v", response.PreferredLocations)
	}
	if response.TargetSeniority != "" {
		t.Fatalf("expected empty target seniority, got %q", response.TargetSeniority)
	}
}

func TestProfileHandlerUpdatesPreferences(t *testing.T) {
	handler := NewProfileHandler(profile.NewService(profile.NewMemoryRepository(), func() time.Time {
		return time.Date(2026, 4, 24, 11, 45, 0, 0, time.UTC)
	}))

	body := map[string]any{
		"preferred_stack":     []string{"Go", " TypeScript ", "go"},
		"remote_only":         true,
		"preferred_locations": []string{"Remote - Europe", " Warsaw "},
		"target_seniority":    " senior ",
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/v1/profile/preferences", bytes.NewReader(payload))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d with body %s", res.Code, res.Body.String())
	}

	var response profilePreferencesResponse
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if !response.RemoteOnly {
		t.Fatalf("expected remote_only to be true")
	}
	if got := response.PreferredStack; len(got) != 2 || got[0] != "Go" || got[1] != "TypeScript" {
		t.Fatalf("unexpected preferred stack: %#v", got)
	}
	if got := response.PreferredLocations; len(got) != 2 || got[0] != "Remote - Europe" || got[1] != "Warsaw" {
		t.Fatalf("unexpected preferred locations: %#v", got)
	}
	if response.TargetSeniority != "senior" {
		t.Fatalf("unexpected target seniority: %q", response.TargetSeniority)
	}
}

func TestProfileHandlerRejectsUnknownFields(t *testing.T) {
	handler := NewProfileHandler(profile.NewService(profile.NewMemoryRepository(), time.Now))

	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/profile/preferences",
		bytes.NewBufferString(`{"preferred_stack":["Go"],"unexpected":true}`),
	)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d with body %s", res.Code, res.Body.String())
	}

	var response map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if response["error"] != "invalid_json" {
		t.Fatalf("expected invalid_json error, got %#v", response)
	}
}
