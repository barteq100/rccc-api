package scoring

import (
	"fmt"
	"strings"
)

const (
	techWeight      = 50
	remoteWeight    = 20
	locationWeight  = 15
	seniorityWeight = 15
)

// Preferences captures the subset of user preferences relevant to deterministic fit scoring.
type Preferences struct {
	PreferredStack     []string
	RemoteOnly         bool
	PreferredLocations []string
	TargetSeniority    string
}

// JobInput captures the subset of job fields evaluated by deterministic fit scoring.
type JobInput struct {
	Title       string
	Location    string
	Remote      bool
	Description string
}

// Result is the deterministic score payload returned to the jobs domain.
type Result struct {
	Score   int
	Reasons []string
}

// Service evaluates deterministic fit scores from user preferences and canonical jobs.
type Service struct{}

// NewService constructs a deterministic fit scoring service.
func NewService() *Service {
	return &Service{}
}

// Evaluate scores a job against the provided preferences and returns explainable reasons.
func (s *Service) Evaluate(job JobInput, preferences Preferences) Result {
	preferences = normalizePreferences(preferences)
	job = normalizeJob(job)

	reasons := make([]string, 0, 4)
	earned := 0
	possible := 0

	if len(preferences.PreferredStack) > 0 {
		possible += techWeight
		matched := matchTerms(preferences.PreferredStack, job.Title+"\n"+job.Description)
		if len(matched) > 0 {
			earned += proportionalWeight(techWeight, len(matched), len(preferences.PreferredStack))
			reasons = append(reasons, fmt.Sprintf("Matched stack keywords: %s.", strings.Join(matched, ", ")))
		} else {
			reasons = append(reasons, "No preferred stack keywords matched.")
		}
	}

	if preferences.RemoteOnly {
		possible += remoteWeight
		if job.Remote {
			earned += remoteWeight
			reasons = append(reasons, "Matches remote-only preference.")
		} else {
			reasons = append(reasons, "Does not match remote-only preference.")
		}
	}

	if len(preferences.PreferredLocations) > 0 {
		possible += locationWeight
		matched := matchTerms(preferences.PreferredLocations, job.Location)
		if len(matched) > 0 {
			earned += locationWeight
			reasons = append(reasons, fmt.Sprintf("Matches preferred locations: %s.", strings.Join(matched, ", ")))
		} else {
			reasons = append(reasons, "Does not match preferred locations.")
		}
	}

	if preferences.TargetSeniority != "" {
		possible += seniorityWeight
		if strings.Contains(job.Title+"\n"+job.Description, strings.ToLower(preferences.TargetSeniority)) {
			earned += seniorityWeight
			reasons = append(reasons, fmt.Sprintf("Matches target seniority: %s.", preferences.TargetSeniority))
		} else {
			reasons = append(reasons, fmt.Sprintf("Does not match target seniority: %s.", preferences.TargetSeniority))
		}
	}

	if possible == 0 {
		return Result{
			Score:   0,
			Reasons: []string{"No profile preferences configured yet."},
		}
	}

	return Result{
		Score:   proportionalWeight(100, earned, possible),
		Reasons: reasons,
	}
}

func normalizePreferences(preferences Preferences) Preferences {
	return Preferences{
		PreferredStack:     normalizeTerms(preferences.PreferredStack),
		RemoteOnly:         preferences.RemoteOnly,
		PreferredLocations: normalizeTerms(preferences.PreferredLocations),
		TargetSeniority:    strings.TrimSpace(preferences.TargetSeniority),
	}
}

func normalizeJob(job JobInput) JobInput {
	return JobInput{
		Title:       strings.ToLower(strings.TrimSpace(job.Title)),
		Location:    strings.ToLower(strings.TrimSpace(job.Location)),
		Remote:      job.Remote,
		Description: strings.ToLower(strings.TrimSpace(job.Description)),
	}
}

func normalizeTerms(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	return normalized
}

func matchTerms(terms []string, haystack string) []string {
	if len(terms) == 0 {
		return nil
	}

	matches := make([]string, 0, len(terms))
	for _, term := range terms {
		if strings.Contains(haystack, strings.ToLower(term)) {
			matches = append(matches, term)
		}
	}

	return matches
}

func proportionalWeight(total, matched, available int) int {
	if matched <= 0 || available <= 0 || total <= 0 {
		return 0
	}

	return (matched*total + available/2) / available
}
