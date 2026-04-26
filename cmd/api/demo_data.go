package main

import (
	"context"
	"time"

	"github.com/barteq100/rccc-api/internal/applications"
	"github.com/barteq100/rccc-api/internal/jobs"
	"github.com/barteq100/rccc-api/internal/profile"
)

func seedDemoRuntime(
	ctx context.Context,
	jobRepo *jobs.MemoryRepository,
	profileRepo *profile.MemoryRepository,
	applicationsRepo *applications.MemoryRepository,
) error {
	for _, job := range demoJobs() {
		if err := jobRepo.Upsert(ctx, job); err != nil {
			return err
		}
	}

	if _, err := profileRepo.Save(ctx, demoPreferences()); err != nil {
		return err
	}

	for _, application := range demoApplications() {
		if err := applicationsRepo.Upsert(ctx, application); err != nil {
			return err
		}
	}

	return nil
}

func demoJobs() []jobs.Job {
	return []jobs.Job{
		{
			ID:          "job-1001",
			Title:       "Senior Go Platform Engineer",
			Company:     "Northstar Labs",
			Location:    "Remote - Europe",
			Remote:      true,
			Description: "Build Go platform services, internal tooling, and distributed systems workflows.",
			Source:      "greenhouse",
			SourceURL:   "https://jobs.example.com/northstar/senior-go-platform-engineer",
			PostedAt:    time.Date(2026, 4, 25, 9, 0, 0, 0, time.UTC),
			IngestedAt:  time.Date(2026, 4, 25, 9, 30, 0, 0, time.UTC),
			CreatedAt:   time.Date(2026, 4, 25, 9, 30, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 4, 25, 9, 30, 0, 0, time.UTC),
		},
		{
			ID:          "job-1002",
			Title:       "Staff Platform Engineer",
			Company:     "Orbit Commerce",
			Location:    "Warsaw",
			Remote:      false,
			Description: "Lead platform engineering and developer experience improvements.",
			Source:      "lever",
			SourceURL:   "https://jobs.example.com/orbit/staff-platform-engineer",
			PostedAt:    time.Date(2026, 4, 24, 13, 0, 0, 0, time.UTC),
			IngestedAt:  time.Date(2026, 4, 24, 13, 15, 0, 0, time.UTC),
			CreatedAt:   time.Date(2026, 4, 24, 13, 15, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 4, 24, 13, 15, 0, 0, time.UTC),
		},
		{
			ID:          "job-1003",
			Title:       "Senior Backend Engineer",
			Company:     "Cloudline",
			Location:    "Remote - Poland",
			Remote:      true,
			Description: "Build backend APIs in Go with PostgreSQL and platform automation.",
			Source:      "greenhouse",
			SourceURL:   "https://jobs.example.com/cloudline/senior-backend-engineer",
			PostedAt:    time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC),
			IngestedAt:  time.Date(2026, 4, 23, 10, 20, 0, 0, time.UTC),
			CreatedAt:   time.Date(2026, 4, 23, 10, 20, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 4, 23, 10, 20, 0, 0, time.UTC),
		},
		{
			ID:          "job-1004",
			Title:       "Engineering Manager",
			Company:     "Summit AI",
			Location:    "Berlin",
			Remote:      false,
			Description: "Manage engineering teams and roadmap delivery for product initiatives.",
			Source:      "lever",
			SourceURL:   "https://jobs.example.com/summit/engineering-manager",
			PostedAt:    time.Date(2026, 4, 22, 15, 0, 0, 0, time.UTC),
			IngestedAt:  time.Date(2026, 4, 22, 15, 10, 0, 0, time.UTC),
			CreatedAt:   time.Date(2026, 4, 22, 15, 10, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 4, 22, 15, 10, 0, 0, time.UTC),
		},
	}
}

func demoPreferences() profile.Preferences {
	return profile.Preferences{
		ID:                 1,
		PreferredStack:     []string{"Go", "Platform", "PostgreSQL"},
		RemoteOnly:         true,
		PreferredLocations: []string{"Europe", "Poland"},
		TargetSeniority:    "senior",
		CreatedAt:          time.Date(2026, 4, 25, 8, 0, 0, 0, time.UTC),
		UpdatedAt:          time.Date(2026, 4, 25, 8, 0, 0, 0, time.UTC),
	}
}

func demoApplications() []applications.Application {
	appliedAt := time.Date(2026, 4, 25, 11, 0, 0, 0, time.UTC)

	return []applications.Application{
		{
			JobID:           "job-1001",
			Status:          applications.StatusSaved,
			SavedAt:         time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC),
			StatusChangedAt: time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC),
			CreatedAt:       time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC),
		},
		{
			JobID:           "job-1003",
			Status:          applications.StatusApplied,
			SavedAt:         time.Date(2026, 4, 24, 16, 0, 0, 0, time.UTC),
			AppliedAt:       &appliedAt,
			StatusChangedAt: appliedAt,
			CreatedAt:       time.Date(2026, 4, 24, 16, 0, 0, 0, time.UTC),
			UpdatedAt:       appliedAt,
		},
		{
			JobID:           "job-1002",
			Status:          applications.StatusInterview,
			SavedAt:         time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC),
			AppliedAt:       &appliedAt,
			StatusChangedAt: time.Date(2026, 4, 26, 9, 0, 0, 0, time.UTC),
			CreatedAt:       time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2026, 4, 26, 9, 0, 0, 0, time.UTC),
		},
	}
}
