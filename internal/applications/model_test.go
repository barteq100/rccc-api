package applications

import "testing"

func TestStatusValid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{name: "saved", status: StatusSaved, want: true},
		{name: "applied", status: StatusApplied, want: true},
		{name: "interview", status: StatusInterview, want: true},
		{name: "offer", status: StatusOffer, want: true},
		{name: "rejected", status: StatusRejected, want: true},
		{name: "invalid", status: Status("draft"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
