package google

import (
	"context"
	"testing"

	directory "google.golang.org/api/admin/directory/v1"
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

func TestConvertUserPhones(t *testing.T) {
	tests := []struct {
		name            string
		phones          interface{}
		wantPhone       string
		wantPhoneMobile string
	}{
		{
			name: "only mobile phone - should not duplicate into Phone",
			phones: []interface{}{
				map[string]interface{}{"type": "mobile", "value": "6971234567"},
			},
			wantPhone:       "",
			wantPhoneMobile: "6971234567",
		},
		{
			name: "only work phone",
			phones: []interface{}{
				map[string]interface{}{"type": "work", "value": "2101234567"},
			},
			wantPhone:       "2101234567",
			wantPhoneMobile: "",
		},
		{
			name: "work and mobile both present",
			phones: []interface{}{
				map[string]interface{}{"type": "work", "value": "2101234567"},
				map[string]interface{}{"type": "mobile", "value": "6971234567"},
			},
			wantPhone:       "2101234567",
			wantPhoneMobile: "6971234567",
		},
		{
			name: "only a non-work, non-mobile type falls back to it",
			phones: []interface{}{
				map[string]interface{}{"type": "home", "value": "2109999999"},
			},
			wantPhone:       "2109999999",
			wantPhoneMobile: "",
		},
		{
			name:            "no phones",
			phones:          nil,
			wantPhone:       "",
			wantPhoneMobile: "",
		},
	}

	d := NewDirectoryClient(nil, "example.com", CompanyConfig{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &directory.User{PrimaryEmail: "alice@example.com", Phones: tt.phones}
			user := d.convertUser(u)

			if user.Phone != tt.wantPhone {
				t.Errorf("Phone = %q, want %q", user.Phone, tt.wantPhone)
			}
			if user.PhoneMobile != tt.wantPhoneMobile {
				t.Errorf("PhoneMobile = %q, want %q", user.PhoneMobile, tt.wantPhoneMobile)
			}
		})
	}
}

func TestConvertUserPhoneLabel(t *testing.T) {
	tests := []struct {
		name           string
		phones         interface{}
		wantPhone      string
		wantPhoneLabel string
		wantIsInternal bool
	}{
		{
			name: "work phone gets the default label",
			phones: []interface{}{
				map[string]interface{}{"type": "work", "value": "2101234567"},
			},
			wantPhone:      "2101234567",
			wantPhoneLabel: "T",
			wantIsInternal: false,
		},
		{
			name: "custom phone with internal customType gets the internal label",
			phones: []interface{}{
				map[string]interface{}{"type": "custom", "customType": "Internal", "value": "1234"},
			},
			wantPhone:      "1234",
			wantPhoneLabel: "Ext.",
			wantIsInternal: true,
		},
		{
			name: "customType matching is case-insensitive and allows extra words",
			phones: []interface{}{
				map[string]interface{}{"type": "custom", "customType": "internal extension", "value": "5678"},
			},
			wantPhone:      "5678",
			wantPhoneLabel: "Ext.",
			wantIsInternal: true,
		},
		{
			name: "custom phone without an internal customType is not labeled internal",
			phones: []interface{}{
				map[string]interface{}{"type": "custom", "customType": "fax line", "value": "2109999999"},
			},
			wantPhone:      "2109999999",
			wantPhoneLabel: "T",
			wantIsInternal: false,
		},
		{
			name: "explicit work phone wins over a later internal custom phone",
			phones: []interface{}{
				map[string]interface{}{"type": "work", "value": "2101234567"},
				map[string]interface{}{"type": "custom", "customType": "internal", "value": "1234"},
			},
			wantPhone:      "2101234567",
			wantPhoneLabel: "T",
			wantIsInternal: false,
		},
		{
			name:           "no phones - no label",
			phones:         nil,
			wantPhone:      "",
			wantPhoneLabel: "",
			wantIsInternal: false,
		},
	}

	d := NewDirectoryClient(nil, "example.com", CompanyConfig{
		PhoneLabel:         "T",
		InternalPhoneLabel: "Ext.",
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &directory.User{PrimaryEmail: "alice@example.com", Phones: tt.phones}
			user := d.convertUser(u)

			if user.Phone != tt.wantPhone {
				t.Errorf("Phone = %q, want %q", user.Phone, tt.wantPhone)
			}
			if user.PhoneLabel != tt.wantPhoneLabel {
				t.Errorf("PhoneLabel = %q, want %q", user.PhoneLabel, tt.wantPhoneLabel)
			}
			if user.PhoneIsInternal != tt.wantIsInternal {
				t.Errorf("PhoneIsInternal = %v, want %v", user.PhoneIsInternal, tt.wantIsInternal)
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
