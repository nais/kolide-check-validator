package main

import (
	kac "github.com/nais/kolide-check-validator/pkg/kolide-api-client"
	"testing"
)

func Test_hasSeverityTag(t *testing.T) {
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
			if got := hasSeverityTag(tt.check); got != tt.hasSeverityTag {
				t.Errorf("hasSeverityTag(%v) = %v, expected %v", tt.check.Tags, got, tt.hasSeverityTag)
			}
		})
	}
}
