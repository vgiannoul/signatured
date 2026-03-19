package template

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
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

// Load reads a signature template from the specified path.
// Supports local files and Google Cloud Storage URLs (gs:// or https://storage.googleapis.com/).
func Load(path string) (*Template, error) {
	var content []byte
	var err error

	if isGCSPath(path) {
		content, err = loadFromGCS(context.Background(), path)
		if err != nil {
			return nil, fmt.Errorf("failed to load template from GCS: %w", err)
		}
	} else {
		content, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read template file: %w", err)
		}
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

// isGCSPath checks if the path is a Google Cloud Storage URL.
func isGCSPath(path string) bool {
	return strings.HasPrefix(path, "gs://") ||
		strings.Contains(path, "storage.googleapis.com")
}

// parseGCSURL extracts bucket and object path from a GCS URL.
// Supports both gs:// and https://storage.googleapis.com/ formats.
func parseGCSURL(url string) (bucket, object string, err error) {
	// gs://bucket-name/path/to/object.md
	if strings.HasPrefix(url, "gs://") {
		url = strings.TrimPrefix(url, "gs://")
		parts := strings.SplitN(url, "/", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid GCS URL format: expected gs://bucket/object")
		}
		return parts[0], parts[1], nil
	}

	// https://storage.googleapis.com/bucket-name/path/to/object.md
	if strings.Contains(url, "storage.googleapis.com") {
		// Remove protocol and split by /
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimPrefix(url, "http://")
		parts := strings.Split(url, "/")

		// Find "storage.googleapis.com" and extract bucket/object after it
		for i, part := range parts {
			if part == "storage.googleapis.com" {
				if i+2 >= len(parts) {
					return "", "", fmt.Errorf("invalid GCS HTTPS URL format")
				}
				bucket = parts[i+1]
				object = strings.Join(parts[i+2:], "/")
				return bucket, object, nil
			}
		}
	}

	return "", "", fmt.Errorf("unrecognized GCS URL format: %s", url)
}

// loadFromGCS downloads a template from Google Cloud Storage.
// Uses Application Default Credentials (ADC) for authentication.
func loadFromGCS(ctx context.Context, gcsURL string) ([]byte, error) {
	bucket, object, err := parseGCSURL(gcsURL)
	if err != nil {
		return nil, err
	}

	// Create GCS client (uses Application Default Credentials)
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	defer client.Close()

	// Open object reader
	reader, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read object gs://%s/%s: %w", bucket, object, err)
	}
	defer reader.Close()

	// Read all content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content from GCS: %w", err)
	}

	return content, nil
}

// Validate checks if the template file exists and can be parsed.
func Validate(path string) error {
	_, err := Load(path)
	return err
}
