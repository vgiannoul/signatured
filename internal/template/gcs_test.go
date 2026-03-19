package template

import (
	"testing"
)

func TestIsGCSPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "gs:// URL",
			path:     "gs://my-bucket/templates/signatured.md",
			expected: true,
		},
		{
			name:     "HTTPS storage.googleapis.com URL",
			path:     "https://storage.googleapis.com/my-bucket/templates/signatured.md",
			expected: true,
		},
		{
			name:     "HTTP storage.googleapis.com URL",
			path:     "http://storage.googleapis.com/my-bucket/templates/signatured.md",
			expected: true,
		},
		{
			name:     "local file path",
			path:     "./templates/signatured.md",
			expected: false,
		},
		{
			name:     "absolute file path",
			path:     "/var/templates/signatured.md",
			expected: false,
		},
		{
			name:     "other HTTPS URL",
			path:     "https://example.com/templates/signatured.md",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGCSPath(tt.path)
			if result != tt.expected {
				t.Errorf("isGCSPath(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestParseGCSURL(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedBucket string
		expectedObject string
		expectError    bool
	}{
		{
			name:           "gs:// URL with simple path",
			url:            "gs://my-bucket/signatured.md",
			expectedBucket: "my-bucket",
			expectedObject: "signatured.md",
			expectError:    false,
		},
		{
			name:           "gs:// URL with nested path",
			url:            "gs://my-org-templates/templates/departments/engineering.md",
			expectedBucket: "my-org-templates",
			expectedObject: "templates/departments/engineering.md",
			expectError:    false,
		},
		{
			name:           "HTTPS storage.googleapis.com URL",
			url:            "https://storage.googleapis.com/my-bucket/signatured.md",
			expectedBucket: "my-bucket",
			expectedObject: "signatured.md",
			expectError:    false,
		},
		{
			name:           "HTTPS storage.googleapis.com URL with nested path",
			url:            "https://storage.googleapis.com/my-bucket/path/to/signatured.md",
			expectedBucket: "my-bucket",
			expectedObject: "path/to/signatured.md",
			expectError:    false,
		},
		{
			name:           "HTTP storage.googleapis.com URL",
			url:            "http://storage.googleapis.com/my-bucket/signatured.md",
			expectedBucket: "my-bucket",
			expectedObject: "signatured.md",
			expectError:    false,
		},
		{
			name:        "gs:// URL without object",
			url:         "gs://my-bucket",
			expectError: true,
		},
		{
			name:        "gs:// URL without bucket",
			url:         "gs://",
			expectError: true,
		},
		{
			name:        "invalid HTTPS URL",
			url:         "https://storage.googleapis.com/my-bucket",
			expectError: true,
		},
		{
			name:        "non-GCS URL",
			url:         "https://example.com/file.md",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, object, err := parseGCSURL(tt.url)

			if tt.expectError {
				if err == nil {
					t.Errorf("parseGCSURL(%q) expected error but got none", tt.url)
				}
				return
			}

			if err != nil {
				t.Errorf("parseGCSURL(%q) unexpected error: %v", tt.url, err)
				return
			}

			if bucket != tt.expectedBucket {
				t.Errorf("parseGCSURL(%q) bucket = %q, expected %q", tt.url, bucket, tt.expectedBucket)
			}

			if object != tt.expectedObject {
				t.Errorf("parseGCSURL(%q) object = %q, expected %q", tt.url, object, tt.expectedObject)
			}
		})
	}
}
