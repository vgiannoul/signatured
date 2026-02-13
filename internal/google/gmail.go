package google

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// GmailClient handles updating Gmail signatures.
// It creates per-user authenticated clients to update each user's signature.
type GmailClient struct {
	credentialsPath string
}

// NewGmailClient creates a new Gmail API client.
func NewGmailClient(credentialsPath string) *GmailClient {
	return &GmailClient{
		credentialsPath: credentialsPath,
	}
}

// UpdateSignature updates the signature for a user's primary email address.
// It impersonates the user to update their own signature.
func (g *GmailClient) UpdateSignature(ctx context.Context, userEmail, signatureHTML string) error {
	// Create a Gmail service impersonating this specific user
	service, err := g.createUserGmailService(ctx, userEmail)
	if err != nil {
		return fmt.Errorf("failed to create Gmail service for %s: %w", userEmail, err)
	}

	// The sendAs address is typically the user's primary email
	sendAsEmail := userEmail

	// Create the signature update request
	sendAs := &gmail.SendAs{
		Signature: signatureHTML,
	}

	// Update the signature with retry logic for rate limiting
	err = g.retryWithBackoff(ctx, func() error {
		_, err := service.Users.Settings.SendAs.
			Patch(userEmail, sendAsEmail, sendAs).
			Context(ctx).
			Do()
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to update signature for %s: %w", userEmail, err)
	}

	return nil
}

// createUserGmailService creates a Gmail service client impersonating a specific user.
func (g *GmailClient) createUserGmailService(ctx context.Context, userEmail string) (*gmail.Service, error) {
	// Read the service account credentials file
	credBytes, err := os.ReadFile(g.credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	// Parse the service account credentials with Gmail scope
	config, err := google.JWTConfigFromJSON(credBytes, GmailSettingsBasicScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	// Set the subject for domain-wide delegation to this specific user
	config.Subject = userEmail

	// Create HTTP client with the user-specific service account credentials
	httpClient := config.Client(ctx)

	// Initialize Gmail API service
	service, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	return service, nil
}

// retryWithBackoff implements exponential backoff retry for API rate limiting.
func (g *GmailClient) retryWithBackoff(ctx context.Context, fn func() error) error {
	maxRetries := 5
	baseDelay := 1 * time.Second
	maxDelay := 32 * time.Second

	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		// Check if the error is a rate limit error
		if apiErr, ok := err.(*googleapi.Error); ok {
			if apiErr.Code == 429 || apiErr.Code == 503 {
				// Calculate delay with exponential backoff
				delay := baseDelay * time.Duration(1<<uint(attempt))
				if delay > maxDelay {
					delay = maxDelay
				}

				// Wait before retrying
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
					continue
				}
			}
		}

		// For non-rate-limit errors, return immediately
		return err
	}

	return fmt.Errorf("max retries exceeded: %w", err)
}
