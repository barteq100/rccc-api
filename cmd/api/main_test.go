package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServerSeedsDemoDataWhenEnabled(t *testing.T) {
	handler, err := newServer(t.Context(), serverOptions{SeedDemoData: true})
	if err != nil {
		t.Fatalf("newServer returned error: %v", err)
	}

	jobsReq := httptest.NewRequest(http.MethodGet, "/v1/jobs?page_size=10", nil)
	jobsRes := httptest.NewRecorder()
	handler.ServeHTTP(jobsRes, jobsReq)

	if jobsRes.Code != http.StatusOK {
		t.Fatalf("expected jobs status 200, got %d with body %s", jobsRes.Code, jobsRes.Body.String())
	}

	var jobsResponse struct {
		Total int `json:"total"`
		Items []struct {
			ID           string   `json:"id"`
			Score        int      `json:"score"`
			ScoreReasons []string `json:"score_reasons"`
		} `json:"items"`
	}
	if err := json.Unmarshal(jobsRes.Body.Bytes(), &jobsResponse); err != nil {
		t.Fatalf("Unmarshal jobs response returned error: %v", err)
	}
	if jobsResponse.Total != 4 {
		t.Fatalf("expected 4 demo jobs, got %d", jobsResponse.Total)
	}
	if len(jobsResponse.Items) != 4 {
		t.Fatalf("expected 4 listed jobs, got %d", len(jobsResponse.Items))
	}
	if jobsResponse.Items[0].ID != "job-1001" {
		t.Fatalf("expected most recent demo job first, got %#v", jobsResponse.Items[0])
	}
	if jobsResponse.Items[0].Score != 83 {
		t.Fatalf("expected seeded profile score of 83, got %d", jobsResponse.Items[0].Score)
	}
	if len(jobsResponse.Items[0].ScoreReasons) == 0 {
		t.Fatalf("expected score reasons for seeded job")
	}

	profileReq := httptest.NewRequest(http.MethodGet, "/v1/profile/preferences", nil)
	profileRes := httptest.NewRecorder()
	handler.ServeHTTP(profileRes, profileReq)

	if profileRes.Code != http.StatusOK {
		t.Fatalf("expected profile status 200, got %d with body %s", profileRes.Code, profileRes.Body.String())
	}

	var profileResponse struct {
		PreferredStack     []string `json:"preferred_stack"`
		RemoteOnly         bool     `json:"remote_only"`
		PreferredLocations []string `json:"preferred_locations"`
		TargetSeniority    string   `json:"target_seniority"`
	}
	if err := json.Unmarshal(profileRes.Body.Bytes(), &profileResponse); err != nil {
		t.Fatalf("Unmarshal profile response returned error: %v", err)
	}
	if !profileResponse.RemoteOnly {
		t.Fatalf("expected seeded profile to be remote_only")
	}
	if len(profileResponse.PreferredStack) != 3 {
		t.Fatalf("expected seeded preferred stack, got %#v", profileResponse.PreferredStack)
	}
	if profileResponse.TargetSeniority != "senior" {
		t.Fatalf("expected seeded target seniority, got %q", profileResponse.TargetSeniority)
	}

	applicationsReq := httptest.NewRequest(http.MethodGet, "/v1/applications", nil)
	applicationsRes := httptest.NewRecorder()
	handler.ServeHTTP(applicationsRes, applicationsReq)

	if applicationsRes.Code != http.StatusOK {
		t.Fatalf("expected applications status 200, got %d with body %s", applicationsRes.Code, applicationsRes.Body.String())
	}

	var applicationsResponse struct {
		Items []struct {
			Job struct {
				ID string `json:"id"`
			} `json:"job"`
			Status string `json:"status"`
		} `json:"items"`
	}
	if err := json.Unmarshal(applicationsRes.Body.Bytes(), &applicationsResponse); err != nil {
		t.Fatalf("Unmarshal applications response returned error: %v", err)
	}
	if len(applicationsResponse.Items) != 3 {
		t.Fatalf("expected 3 seeded applications, got %d", len(applicationsResponse.Items))
	}
	if applicationsResponse.Items[0].Job.ID != "job-1002" || applicationsResponse.Items[0].Status != "interview" {
		t.Fatalf("expected seeded interview application first, got %#v", applicationsResponse.Items[0])
	}
}

func TestNewServerStartsEmptyWhenDemoDataDisabled(t *testing.T) {
	handler, err := newServer(t.Context(), serverOptions{})
	if err != nil {
		t.Fatalf("newServer returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/jobs?page_size=10", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected jobs status 200, got %d with body %s", res.Code, res.Body.String())
	}

	var response struct {
		Total int        `json:"total"`
		Items []struct{} `json:"items"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal response returned error: %v", err)
	}
	if response.Total != 0 || len(response.Items) != 0 {
		t.Fatalf("expected empty runtime without demo data, got %#v", response)
	}
}
