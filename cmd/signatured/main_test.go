package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vgiannoul/signatured/internal/models"
	"github.com/vgiannoul/signatured/internal/template"
)

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		email string
		want  string
	}{
		{"admin@example.com", "example.com"},
		{"no-at-sign", ""},
		{"", ""},
		{"a@b@example.com", "example.com"},
	}

	for _, tt := range tests {
		if got := extractDomain(tt.email); got != tt.want {
			t.Errorf("extractDomain(%q) = %q, want %q", tt.email, got, tt.want)
		}
	}
}

func TestGetEnvSlice(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{"unset", "", nil},
		{"single value", "alice@example.com", []string{"alice@example.com"}},
		{"multiple values", "alice@example.com,bob@example.com", []string{"alice@example.com", "bob@example.com"}},
		{"trims whitespace", " alice@example.com , bob@example.com ", []string{"alice@example.com", "bob@example.com"}},
		{"drops empty entries", "alice@example.com,,bob@example.com,", []string{"alice@example.com", "bob@example.com"}},
	}

	const key = "SIGNATURED_TEST_ENV_SLICE"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(key, tt.value)
			got := getEnvSlice(key)
			if len(got) != len(tt.want) {
				t.Fatalf("getEnvSlice(%q) = %v, want %v", tt.value, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getEnvSlice(%q)[%d] = %q, want %q", tt.value, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestValidateApplyTarget(t *testing.T) {
	tests := []struct {
		name      string
		userEmail string
		orgUnit   string
		applyAll  bool
		wantErr   bool
	}{
		{"none set", "", "", false, true},
		{"user only", "alice@example.com", "", false, false},
		{"org-unit only", "", "/Engineering", false, false},
		{"all only", "", "", true, false},
		{"user and org-unit", "alice@example.com", "/Engineering", false, true},
		{"all three", "alice@example.com", "/Engineering", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateApplyTarget(tt.userEmail, tt.orgUnit, tt.applyAll)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateApplyTarget(%q, %q, %v) error = %v, wantErr %v",
					tt.userEmail, tt.orgUnit, tt.applyAll, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePreviewTarget(t *testing.T) {
	tests := []struct {
		name      string
		sample    bool
		userEmail string
		wantErr   bool
	}{
		{"neither set", false, "", true},
		{"sample only", true, "", false},
		{"user only", false, "alice@example.com", false},
		{"both set", true, "alice@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePreviewTarget(tt.sample, tt.userEmail)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePreviewTarget(%v, %q) error = %v, wantErr %v",
					tt.sample, tt.userEmail, err, tt.wantErr)
			}
		})
	}
}

// fakeApplier is a test double for signatureApplier that records calls and
// can be configured to fail for specific users or track concurrency.
type fakeApplier struct {
	mu          sync.Mutex
	calls       []string
	failFor     map[string]bool
	inFlight    int32
	maxInFlight int32
	delay       time.Duration
}

func (f *fakeApplier) UpdateSignature(ctx context.Context, userEmail, signatureHTML string) error {
	current := atomic.AddInt32(&f.inFlight, 1)
	defer atomic.AddInt32(&f.inFlight, -1)

	for {
		max := atomic.LoadInt32(&f.maxInFlight)
		if current <= max || atomic.CompareAndSwapInt32(&f.maxInFlight, max, current) {
			break
		}
	}

	if f.delay > 0 {
		time.Sleep(f.delay)
	}

	f.mu.Lock()
	f.calls = append(f.calls, userEmail)
	shouldFail := f.failFor[userEmail]
	f.mu.Unlock()

	if shouldFail {
		return fmt.Errorf("simulated failure for %s", userEmail)
	}
	return nil
}

func loadTestTemplate(t *testing.T) *template.Template {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "signature.md")
	if err := os.WriteFile(path, []byte("{{firstName}} {{lastName}}"), 0644); err != nil {
		t.Fatalf("failed to write test template: %v", err)
	}
	tmpl, err := template.Load(path)
	if err != nil {
		t.Fatalf("failed to load test template: %v", err)
	}
	return tmpl
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError + 1}))
}

func usersWithEmails(emails ...string) []*models.User {
	users := make([]*models.User, 0, len(emails))
	for _, email := range emails {
		users = append(users, &models.User{Email: email, FirstName: "Test", LastName: "User"})
	}
	return users
}

func TestProcessUsersAllSucceed(t *testing.T) {
	applier := &fakeApplier{}
	tmpl := loadTestTemplate(t)
	users := usersWithEmails("a@example.com", "b@example.com", "c@example.com")

	err := processUsers(context.Background(), discardLogger(), tmpl, applier, users, 10, false, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(applier.calls) != 3 {
		t.Errorf("expected 3 UpdateSignature calls, got %d", len(applier.calls))
	}
}

func TestProcessUsersReportsFailures(t *testing.T) {
	applier := &fakeApplier{failFor: map[string]bool{"b@example.com": true}}
	tmpl := loadTestTemplate(t)
	users := usersWithEmails("a@example.com", "b@example.com", "c@example.com")

	err := processUsers(context.Background(), discardLogger(), tmpl, applier, users, 10, false, nil)
	if err == nil {
		t.Fatal("expected an error because one user failed, got nil")
	}
	if !strings.Contains(err.Error(), "1 users failed to update") {
		t.Errorf("unexpected error message: %v", err)
	}
	if len(applier.calls) != 3 {
		t.Errorf("expected all 3 users to still be attempted, got %d calls", len(applier.calls))
	}
}

func TestProcessUsersDryRunSkipsUpdates(t *testing.T) {
	applier := &fakeApplier{}
	tmpl := loadTestTemplate(t)
	users := usersWithEmails("a@example.com", "b@example.com")

	err := processUsers(context.Background(), discardLogger(), tmpl, applier, users, 10, true, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(applier.calls) != 0 {
		t.Errorf("dry run should not call UpdateSignature, got %d calls", len(applier.calls))
	}
}

func TestProcessUsersExcludesUsers(t *testing.T) {
	applier := &fakeApplier{}
	tmpl := loadTestTemplate(t)
	users := usersWithEmails("a@example.com", "b@example.com", "c@example.com")

	err := processUsers(context.Background(), discardLogger(), tmpl, applier, users, 10, false,
		[]string{" B@Example.com "}) // exercises trimming and case-insensitivity too
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(applier.calls) != 2 {
		t.Fatalf("expected 2 UpdateSignature calls (excluded user skipped), got %d: %v", len(applier.calls), applier.calls)
	}
	for _, email := range applier.calls {
		if email == "b@example.com" {
			t.Error("excluded user b@example.com should not have had UpdateSignature called")
		}
	}
}

func TestProcessUsersRespectsConcurrencyLimit(t *testing.T) {
	applier := &fakeApplier{delay: 20 * time.Millisecond}
	tmpl := loadTestTemplate(t)

	emails := make([]string, 20)
	for i := range emails {
		emails[i] = fmt.Sprintf("user%d@example.com", i)
	}
	users := usersWithEmails(emails...)

	const limit = 3
	if err := processUsers(context.Background(), discardLogger(), tmpl, applier, users, limit, false, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if applier.maxInFlight > limit {
		t.Errorf("max concurrent UpdateSignature calls = %d, want <= %d", applier.maxInFlight, limit)
	}
	if applier.maxInFlight == 0 {
		t.Error("expected at least one concurrent call to be recorded")
	}
}
