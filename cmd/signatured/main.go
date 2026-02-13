package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/vgiannoul/signatured/internal/google"
	"github.com/vgiannoul/signatured/internal/models"
	"github.com/vgiannoul/signatured/internal/template"
)

const version = "1.0.0"

var (
	// Global flags
	templatePath    string
	credentialsPath string
	impersonateUser string
	verbose         bool

	// Apply command flags
	dryRun      bool
	concurrency int
	userEmail   string
	orgUnit     string
	applyAll    bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "signatured",
	Short: "Manage Google Workspace email signatures",
	Long: `A CLI tool to manage email signatures for Google Workspace organization members.
Reads signature template from signatured.md, replaces placeholders with user data,
and applies signatures via Google Workspace APIs.`,
	Version: version,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&templatePath, "template", "./signatured.md", "Path to signature template file")
	rootCmd.PersistentFlags().StringVar(&credentialsPath, "credentials", "./credentials.json", "Path to service account credentials")
	rootCmd.PersistentFlags().StringVar(&impersonateUser, "impersonate", "", "User email to impersonate for domain-wide delegation (required)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	// Add subcommands
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(validateCmd)

	// Apply command flags
	applyCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")
	applyCmd.Flags().IntVar(&concurrency, "concurrency", 10, "Number of concurrent API calls")
	applyCmd.Flags().StringVar(&userEmail, "user", "", "Apply to a single user by email")
	applyCmd.Flags().StringVar(&orgUnit, "org-unit", "", "Apply to users in a specific organizational unit")
	applyCmd.Flags().BoolVar(&applyAll, "all", false, "Apply to all users in the domain")
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the signature template",
	Long:  "Validates that the signature template file exists and can be parsed.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := setupLogger(verbose)
		logger.Info("Validating template", "path", templatePath)

		if err := template.Validate(templatePath); err != nil {
			logger.Error("Template validation failed", "error", err)
			return err
		}

		logger.Info("Template is valid")
		return nil
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply signatures to users",
	Long: `Apply email signatures to Google Workspace users.
Specify --user for a single user, --org-unit for an organizational unit,
or --all for the entire domain.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := setupLogger(verbose)

		// Validate flags
		targetCount := 0
		if userEmail != "" {
			targetCount++
		}
		if orgUnit != "" {
			targetCount++
		}
		if applyAll {
			targetCount++
		}

		if targetCount == 0 {
			return fmt.Errorf("must specify one of: --user, --org-unit, or --all")
		}
		if targetCount > 1 {
			return fmt.Errorf("can only specify one of: --user, --org-unit, or --all")
		}

		if impersonateUser == "" {
			return fmt.Errorf("--impersonate is required for domain-wide delegation")
		}

		// Load template
		logger.Info("Loading template", "path", templatePath)
		tmpl, err := template.Load(templatePath)
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}

		// Read template file to show size
		fileInfo, _ := os.Stat(templatePath)
		logger.Info("Template loaded", "size", fmt.Sprintf("%d bytes", fileInfo.Size()))

		// Create Google API client
		ctx := context.Background()
		logger.Info("Authenticating with Google Workspace", "impersonate", impersonateUser)
		client, err := google.NewClient(ctx, credentialsPath, impersonateUser)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		logger.Info("Authentication successful")

		// Create API clients
		directoryClient := google.NewDirectoryClient(client.DirectoryService(), extractDomain(impersonateUser))
		gmailClient := google.NewGmailClient(credentialsPath)

		// Fetch users based on target
		var users []*models.User
		if userEmail != "" {
			logger.Info("Fetching user", "email", userEmail)
			user, err := directoryClient.GetUser(ctx, userEmail)
			if err != nil {
				return err
			}
			users = []*models.User{user}
		} else if orgUnit != "" {
			logger.Info("Fetching users from organizational unit", "orgUnit", orgUnit)
			users, err = directoryClient.ListUsersByOrgUnit(ctx, orgUnit)
			if err != nil {
				return err
			}
		} else {
			logger.Info("Fetching all users in domain")
			users, err = directoryClient.ListUsers(ctx)
			if err != nil {
				return err
			}
		}

		logger.Info("Users found", "count", len(users))

		if dryRun {
			logger.Info("DRY RUN MODE - No changes will be applied")
		}

		// Process users
		return processUsers(ctx, logger, tmpl, gmailClient, users, concurrency, dryRun)
	},
}

// processUsers applies signatures to users with concurrency control.
func processUsers(ctx context.Context, logger *slog.Logger, tmpl *template.Template,
	gmailClient *google.GmailClient, users []*models.User, concurrency int, dryRun bool) error {

	type result struct {
		email   string
		success bool
		err     error
		skipped bool
		reason  string
	}

	results := make(chan result, len(users))
	var wg sync.WaitGroup

	// Semaphore for concurrency control
	sem := make(chan struct{}, concurrency)

	startTime := time.Now()

	logger.Info("Processing users")

	for _, user := range users {
		wg.Add(1)
		go func(u *models.User) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Render signature
			signatureHTML, err := tmpl.Render(u)
			if err != nil {
				results <- result{email: u.Email, success: false, err: err}
				return
			}

			// Skip if dry run
			if dryRun {
				results <- result{email: u.Email, success: true}
				return
			}

			// Update signature
			if err := gmailClient.UpdateSignature(ctx, u.Email, signatureHTML); err != nil {
				results <- result{email: u.Email, success: false, err: err}
				return
			}

			results <- result{email: u.Email, success: true}
		}(user)
	}

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var success, failed, skipped int
	for r := range results {
		if r.skipped {
			skipped++
			logger.Warn("User skipped", "email", r.email, "reason", r.reason)
			fmt.Printf("⊘ %s - skipped (%s)\n", r.email, r.reason)
		} else if r.success {
			success++
			logger.Info("Signature updated", "email", r.email)
			if dryRun {
				fmt.Printf("✓ %s - would update signature (dry run)\n", r.email)
			} else {
				fmt.Printf("✓ %s - signature updated\n", r.email)
			}
		} else {
			failed++
			logger.Error("Signature update failed", "email", r.email, "error", r.err)
			fmt.Printf("✗ %s - error: %v\n", r.email, r.err)
		}
	}

	duration := time.Since(startTime)

	// Print summary
	fmt.Println("\nSummary:")
	fmt.Printf("  Success: %d\n", success)
	fmt.Printf("  Failed:  %d\n", failed)
	fmt.Printf("  Skipped: %d\n", skipped)
	fmt.Printf("  Total:   %d\n", len(users))
	fmt.Printf("  Duration: %.1fs\n", duration.Seconds())

	logger.Info("Processing complete",
		"success", success,
		"failed", failed,
		"skipped", skipped,
		"total", len(users),
		"duration", duration.String(),
	)

	if failed > 0 {
		return fmt.Errorf("%d users failed to update", failed)
	}

	return nil
}

// setupLogger creates a structured logger with the specified verbosity.
func setupLogger(verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	return slog.New(handler)
}

// extractDomain extracts the domain from an email address.
func extractDomain(email string) string {
	for i := len(email) - 1; i >= 0; i-- {
		if email[i] == '@' {
			return email[i+1:]
		}
	}
	return ""
}
