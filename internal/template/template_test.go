package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vgiannoul/signatured/internal/models"
)

func TestLoad(t *testing.T) {
	// Create a temporary template file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "signature.md")

	content := "**{{firstName}} {{lastName}}**\n{{email}}"
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	// Test loading the template
	tmpl, err := Load(templatePath)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	if tmpl.raw != content {
		t.Errorf("Template content mismatch. Got %q, want %q", tmpl.raw, content)
	}
}

func TestLoadNonExistent(t *testing.T) {
	_, err := Load("/nonexistent/template.md")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		template string
		user     *models.User
		contains []string
	}{
		{
			name:     "basic placeholders",
			template: "**{{firstName}} {{lastName}}**\n{{email}}",
			user: &models.User{
				FirstName: "Alice",
				LastName:  "Smith",
				Email:     "alice@example.com",
			},
			contains: []string{"Alice", "Smith", "alice@example.com", "<strong>"},
		},
		{
			name:     "all placeholders",
			template: "{{firstName}} {{lastName}}\n{{jobTitle}}\n{{organization}}\n{{phone}}\n{{email}}\n{{orgUnit}}",
			user: &models.User{
				FirstName:    "Bob",
				LastName:     "Jones",
				JobTitle:     "Engineer",
				Organization: "Acme Corp",
				Phone:        "+1-555-0100",
				Email:        "bob@example.com",
				OrgUnit:      "/Engineering",
			},
			contains: []string{"Bob", "Jones", "Engineer", "Acme Corp", "+1-555-0100", "bob@example.com", "/Engineering"},
		},
		{
			name:     "missing placeholders",
			template: "{{firstName}} {{lastName}}\n{{phone}}",
			user: &models.User{
				FirstName: "Charlie",
				LastName:  "Brown",
				// Phone is missing
			},
			contains: []string{"Charlie", "Brown"},
		},
		{
			name:     "markdown formatting",
			template: "**Bold** and *italic*\n\n---\n\nHorizontal rule above",
			user:     &models.User{},
			contains: []string{"<strong>Bold</strong>", "<em>italic</em>", "<hr>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary template file
			tmpDir := t.TempDir()
			templatePath := filepath.Join(tmpDir, "signature.md")
			if err := os.WriteFile(templatePath, []byte(tt.template), 0644); err != nil {
				t.Fatalf("Failed to create test template: %v", err)
			}

			// Load template
			tmpl, err := Load(templatePath)
			if err != nil {
				t.Fatalf("Failed to load template: %v", err)
			}

			// Render template
			html, err := tmpl.Render(tt.user)
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			// Check for expected content
			for _, want := range tt.contains {
				if !strings.Contains(html, want) {
					t.Errorf("Rendered HTML does not contain %q.\nGot: %s", want, html)
				}
			}
		})
	}
}

func TestReplacePlaceholders(t *testing.T) {
	tmpl := &Template{
		raw: "Hello {{firstName}} {{lastName}}! Email: {{email}}",
	}

	user := &models.User{
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice@example.com",
	}

	content := "Hello {{firstName}} {{lastName}}! Email: {{email}}"
	result := tmpl.replacePlaceholders(user, content)
	expected := "Hello Alice Smith! Email: alice@example.com"

	if result != expected {
		t.Errorf("Placeholder replacement failed.\nGot:  %q\nWant: %q", result, expected)
	}
}

func TestReplacePlaceholdersGracefulDegradation(t *testing.T) {
	tmpl := &Template{
		raw: "{{firstName}} {{lastName}} - {{unknownField}}",
	}

	user := &models.User{
		FirstName: "Bob",
		LastName:  "Jones",
	}

	content := "{{firstName}} {{lastName}} - {{unknownField}}"
	result := tmpl.replacePlaceholders(user, content)
	expected := "Bob Jones - "

	if result != expected {
		t.Errorf("Graceful degradation failed.\nGot:  %q\nWant: %q", result, expected)
	}
}

func TestProcessConditionals(t *testing.T) {
	tests := []struct {
		name     string
		template string
		user     *models.User
		expected string
	}{
		{
			name:     "field present - keep content",
			template: "{{#if phone}}Phone: {{phone}}{{/if}}",
			user: &models.User{
				Phone: "+1-555-0100",
			},
			expected: "Phone: {{phone}}",
		},
		{
			name:     "field missing - remove block",
			template: "{{#if phone}}Phone: {{phone}}{{/if}}",
			user: &models.User{
				// Phone is missing
			},
			expected: "",
		},
		{
			name:     "multiple conditionals - mixed",
			template: "{{#if phone}}📞 {{phone}}\n{{/if}}{{#if email}}✉️ {{email}}{{/if}}",
			user: &models.User{
				Email: "test@example.com",
				// Phone is missing
			},
			expected: "✉️ {{email}}",
		},
		{
			name:     "nested content with newlines",
			template: "{{#if jobTitle}}Title: {{jobTitle}}\nDepartment: Engineering{{/if}}",
			user: &models.User{
				JobTitle: "Software Engineer",
			},
			expected: "Title: {{jobTitle}}\nDepartment: Engineering",
		},
		{
			name:     "all fields present",
			template: "{{#if firstName}}{{firstName}}{{/if}} {{#if lastName}}{{lastName}}{{/if}}",
			user: &models.User{
				FirstName: "John",
				LastName:  "Doe",
			},
			expected: "{{firstName}} {{lastName}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := &Template{
				raw: tt.template,
			}

			result := tmpl.processConditionals(tt.user)

			if result != tt.expected {
				t.Errorf("Conditional processing failed.\nGot:  %q\nWant: %q", result, tt.expected)
			}
		})
	}
}

func TestRenderWithConditionals(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		user        *models.User
		contains    []string
		notContains []string
	}{
		{
			name:     "with phone number",
			template: "{{firstName}} {{lastName}}\n{{#if phone}}📞 {{phone}}{{/if}}",
			user: &models.User{
				FirstName: "Alice",
				LastName:  "Smith",
				Phone:     "+1-555-0100",
			},
			contains:    []string{"Alice Smith", "📞", "+1-555-0100"},
			notContains: []string{},
		},
		{
			name:     "without phone number",
			template: "{{firstName}} {{lastName}}\n{{#if phone}}📞 {{phone}}{{/if}}",
			user: &models.User{
				FirstName: "Bob",
				LastName:  "Jones",
				// Phone is missing
			},
			contains:    []string{"Bob Jones"},
			notContains: []string{"📞"},
		},
		{
			name:     "multiple optional fields",
			template: "{{#if jobTitle}}{{jobTitle}}\n{{/if}}{{#if organization}}{{organization}}\n{{/if}}{{#if phone}}📞 {{phone}}{{/if}}",
			user: &models.User{
				JobTitle: "Engineer",
				// Organization and phone are missing
			},
			contains:    []string{"Engineer"},
			notContains: []string{"📞"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary template file
			tmpDir := t.TempDir()
			templatePath := filepath.Join(tmpDir, "signature.md")
			if err := os.WriteFile(templatePath, []byte(tt.template), 0644); err != nil {
				t.Fatalf("Failed to create test template: %v", err)
			}

			// Load template
			tmpl, err := Load(templatePath)
			if err != nil {
				t.Fatalf("Failed to load template: %v", err)
			}

			// Render template
			html, err := tmpl.Render(tt.user)
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			// Check for expected content
			for _, want := range tt.contains {
				if !strings.Contains(html, want) {
					t.Errorf("Rendered HTML does not contain %q.\nGot: %s", want, html)
				}
			}

			// Check that unwanted content is not present
			for _, unwanted := range tt.notContains {
				if strings.Contains(html, unwanted) {
					t.Errorf("Rendered HTML should not contain %q.\nGot: %s", unwanted, html)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Test valid template
	tmpDir := t.TempDir()
	validPath := filepath.Join(tmpDir, "valid.md")
	if err := os.WriteFile(validPath, []byte("Valid template"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := Validate(validPath); err != nil {
		t.Errorf("Validate failed for valid template: %v", err)
	}

	// Test invalid path
	if err := Validate("/nonexistent/file.md"); err == nil {
		t.Error("Validate should fail for non-existent file")
	}
}
