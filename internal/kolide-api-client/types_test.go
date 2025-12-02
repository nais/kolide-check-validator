package kolide_api_client_test

import (
	"testing"

	kac "github.com/nais/kolide-check-validator/internal/kolide-api-client"
)

func TestCheck_HasSeverityTag(t *testing.T) {
	tests := []struct {
		name           string
		check          kac.Check
		hasSeverityTag bool
	}{
		{
			name: "check with no tags",
			check: kac.Check{
				Tags: nil,
			},
			hasSeverityTag: false,
		},
		{
			name: "check with no severity tags",
			check: kac.Check{
				Tags: []string{"foo", "bar"},
			},
			hasSeverityTag: false,
		},
		{
			name: "check with severity tag",
			check: kac.Check{
				Tags: []string{"info", "foo", "bar"},
			},
			hasSeverityTag: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.check.HasSeverityTag(); got != tt.hasSeverityTag {
				t.Errorf("HasSeverityTag(%v) = %v, expected %v", tt.check.Tags, got, tt.hasSeverityTag)
			}
		})
	}
}
