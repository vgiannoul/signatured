package google

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Required OAuth scopes for the application.
const (
	DirectoryUserReadOnlyScope = "https://www.googleapis.com/auth/admin.directory.user.readonly"
	GmailSettingsBasicScope    = "https://www.googleapis.com/auth/gmail.settings.basic"
)

// Client provides access to Google Workspace APIs.
type Client struct {
	directoryService *directory.Service
	gmailService     *gmail.Service
	impersonateUser  string
}

// NewClient creates a new Google API client using service account credentials.
// The credentialsPath should point to a service account JSON key file.
// The impersonateUser is required for domain-wide delegation (e.g., admin@example.com).
func NewClient(ctx context.Context, credentialsPath, impersonateUser string) (*Client, error) {
	if credentialsPath == "" {
		return nil, fmt.Errorf("credentials path cannot be empty")
	}

	if impersonateUser == "" {
		return nil, fmt.Errorf("impersonate user cannot be empty (required for domain-wide delegation)")
	}

	// Read the service account credentials file
	credBytes, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	// Parse the service account credentials
	config, err := google.JWTConfigFromJSON(credBytes,
		DirectoryUserReadOnlyScope,
		GmailSettingsBasicScope,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	// Set the subject for domain-wide delegation
	config.Subject = impersonateUser

	// Create HTTP client with the service account credentials
	httpClient := config.Client(ctx)

	// Initialize Directory API service
	directoryService, err := directory.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create directory service: %w", err)
	}

	// Initialize Gmail API service
	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create gmail service: %w", err)
	}

	return &Client{
		directoryService: directoryService,
		gmailService:     gmailService,
		impersonateUser:  impersonateUser,
	}, nil
}

// DirectoryService returns the Directory API service for user queries.
func (c *Client) DirectoryService() *directory.Service {
	return c.directoryService
}

// GmailService returns the Gmail API service for signature updates.
func (c *Client) GmailService() *gmail.Service {
	return c.gmailService
}
