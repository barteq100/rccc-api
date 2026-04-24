package scoring

import "testing"

func TestServiceEvaluateReturnsWeightedScoreAndReasons(t *testing.T) {
	service := NewService()

	result := service.Evaluate(JobInput{
		Title:       "Senior Go Platform Engineer",
		Location:    "Remote - Poland",
		Remote:      true,
		Description: "Build Go services and internal tooling.",
	}, Preferences{
		PreferredStack:     []string{"Go", "Kubernetes"},
		RemoteOnly:         true,
		PreferredLocations: []string{"Poland"},
		TargetSeniority:    "senior",
	})

	if result.Score != 75 {
		t.Fatalf("expected weighted score 75, got %d", result.Score)
	}

	expectedReasons := []string{
		"Matched stack keywords: Go.",
		"Matches remote-only preference.",
		"Matches preferred locations: Poland.",
		"Matches target seniority: senior.",
	}
	assertReasons(t, result.Reasons, expectedReasons)
}

func TestServiceEvaluateReturnsZeroWhenNothingMatches(t *testing.T) {
	service := NewService()

	result := service.Evaluate(JobInput{
		Title:       "Onsite Product Designer",
		Location:    "Berlin",
		Remote:      false,
		Description: "Design systems and product interfaces.",
	}, Preferences{
		PreferredStack:     []string{"Go"},
		RemoteOnly:         true,
		PreferredLocations: []string{"Poland"},
		TargetSeniority:    "senior",
	})

	if result.Score != 0 {
		t.Fatalf("expected weighted score 0, got %d", result.Score)
	}

	expectedReasons := []string{
		"No preferred stack keywords matched.",
		"Does not match remote-only preference.",
		"Does not match preferred locations.",
		"Does not match target seniority: senior.",
	}
	assertReasons(t, result.Reasons, expectedReasons)
}

func TestServiceEvaluateReturnsFallbackReasonWithoutConfiguredPreferences(t *testing.T) {
	service := NewService()

	result := service.Evaluate(JobInput{
		Title:       "Backend Engineer",
		Location:    "Remote",
		Remote:      true,
		Description: "Build APIs.",
	}, Preferences{})

	if result.Score != 0 {
		t.Fatalf("expected score 0, got %d", result.Score)
	}

	expectedReasons := []string{"No profile preferences configured yet."}
	assertReasons(t, result.Reasons, expectedReasons)
}

func assertReasons(t *testing.T, actual []string, expected []string) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("expected %d reasons, got %d: %#v", len(expected), len(actual), actual)
	}

	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("expected reason %d to be %q, got %q", index, expected[index], actual[index])
		}
	}
}
