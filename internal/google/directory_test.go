package google

import (
	"context"
	"testing"
)

func TestOrgUnitPathPattern(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		valid bool
	}{
		{"root", "/", true},
		{"nested path", "/Engineering/Backend", true},
		{"with spaces and dashes", "/Sales Team-EU", true},
		{"single quote injection", "/Sales' or orgUnitPath='/", false},
		{"double quote", `/Sales"`, false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := orgUnitPathPattern.MatchString(tt.path)
			if got != tt.valid {
				t.Errorf("orgUnitPathPattern.MatchString(%q) = %v, want %v", tt.path, got, tt.valid)
			}
		})
	}
}

func TestListUsersByOrgUnitRejectsInvalidPath(t *testing.T) {
	// service is intentionally nil: an invalid path must be rejected before
	// the Directory API service is ever touched.
	d := NewDirectoryClient(nil, "example.com", CompanyConfig{})

	_, err := d.ListUsersByOrgUnit(context.Background(), "/Sales' or orgUnitPath='/")
	if err == nil {
		t.Fatal("expected error for org unit path containing a single quote, got nil")
	}
}
