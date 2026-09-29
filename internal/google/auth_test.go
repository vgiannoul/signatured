package google

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewClientRejectsEmptyCredentialsPath(t *testing.T) {
	_, err := NewClient(context.Background(), "", "admin@example.com")
	if err == nil {
		t.Fatal("expected error for empty credentials path, got nil")
	}
	if !strings.Contains(err.Error(), "credentials path cannot be empty") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNewClientRejectsEmptyImpersonateUser(t *testing.T) {
	_, err := NewClient(context.Background(), "/some/path.json", "")
	if err == nil {
		t.Fatal("expected error for empty impersonate user, got nil")
	}
	if !strings.Contains(err.Error(), "impersonate user cannot be empty") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNewClientRejectsMissingCredentialsFile(t *testing.T) {
	_, err := NewClient(context.Background(), "/nonexistent/credentials.json", "admin@example.com")
	if err == nil {
		t.Fatal("expected error for missing credentials file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read credentials file") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNewClientRejectsInvalidCredentialsJSON(t *testing.T) {
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, "credentials.json")
	if err := os.WriteFile(credPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write test credentials file: %v", err)
	}

	_, err := NewClient(context.Background(), credPath, "admin@example.com")
	if err == nil {
		t.Fatal("expected error for invalid credentials JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse credentials") {
		t.Errorf("unexpected error message: %v", err)
	}
}
