package template

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/vgiannoul/signatured/internal/models"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// Template represents a signature template with markdown content.
type Template struct {
	raw      string
	markdown goldmark.Markdown
}

// Load reads a signature template from the specified file path.
func Load(path string) (*Template, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // Allow raw HTML in markdown
		),
	)

	return &Template{
		raw:      string(content),
		markdown: md,
	}, nil
}

// Render converts the template to HTML and replaces placeholders with user data.
func (t *Template) Render(user *models.User) (string, error) {
	// First, process conditional blocks based on user data
	content := t.processConditionals(user)

	// Then, replace remaining placeholders in the markdown content
	content = t.replacePlaceholders(user, content)

	// Convert markdown to HTML
	var buf bytes.Buffer
	if err := t.markdown.Convert([]byte(content), &buf); err != nil {
		return "", fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	html := buf.String()

	// Clean up the HTML (remove unnecessary whitespace)
	html = strings.TrimSpace(html)

	return html, nil
}

// processConditionals processes {{#if field}}...{{/if}} blocks.
// If the field has no value, the entire block is removed.
func (t *Template) processConditionals(user *models.User) string {
	content := t.raw
	placeholders := user.PlaceholderData()

	// Pattern matches {{#if fieldName}}content{{/if}}
	// (?s) flag makes . match newlines as well
	// Non-greedy match (.*?) to handle multiple conditionals correctly
	pattern := regexp.MustCompile(`(?s)\{\{#if\s+(\w+)\}\}(.*?)\{\{/if\}\}`)

	content = pattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the field name and content between the tags
		submatches := pattern.FindStringSubmatch(match)
		if len(submatches) != 3 {
			return match
		}

		fieldName := strings.TrimSpace(submatches[1])
		innerContent := submatches[2]

		// Check if the field has a value
		value, ok := placeholders[fieldName]
		if !ok || value == "" {
			// Field is missing or empty, remove the entire block
			return ""
		}

		// Special case: treat top-level orgUnit ("/") as empty
		if fieldName == "orgUnit" && value == "/" {
			return ""
		}

		// Field has a value, keep the inner content
		return innerContent
	})

	return content
}

// replacePlaceholders replaces {{placeholder}} syntax with actual user data.
func (t *Template) replacePlaceholders(user *models.User, content string) string {
	placeholders := user.PlaceholderData()

	// Pattern matches {{placeholder}} syntax
	pattern := regexp.MustCompile(`\{\{(\w+)\}\}`)

	content = pattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the key from {{key}}
		key := strings.TrimSpace(match[2 : len(match)-2])

		// Look up the value in the user data
		if value, ok := placeholders[key]; ok {
			return value
		}

		// If no value found, return empty string (graceful degradation)
		return ""
	})

	return content
}

// Validate checks if the template file exists and can be parsed.
func Validate(path string) error {
	_, err := Load(path)
	return err
}
