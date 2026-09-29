package google

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/api/googleapi"
)

func TestRetryWithBackoffSucceedsImmediately(t *testing.T) {
	calls := 0
	err := retryWithBackoff(context.Background(), func() error {
		calls++
		return nil
	}, 5, time.Millisecond, 5*time.Millisecond)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryWithBackoffRetriesOnRateLimit(t *testing.T) {
	calls := 0
	err := retryWithBackoff(context.Background(), func() error {
		calls++
		if calls < 3 {
			return &googleapi.Error{Code: 429}
		}
		return nil
	}, 5, time.Millisecond, 5*time.Millisecond)

	if err != nil {
		t.Fatalf("expected no error after retries, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetryWithBackoffRetriesOn503(t *testing.T) {
	calls := 0
	err := retryWithBackoff(context.Background(), func() error {
		calls++
		if calls < 2 {
			return &googleapi.Error{Code: 503}
		}
		return nil
	}, 5, time.Millisecond, 5*time.Millisecond)

	if err != nil {
		t.Fatalf("expected no error after retries, got %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestRetryWithBackoffDoesNotRetryOtherErrors(t *testing.T) {
	calls := 0
	wantErr := &googleapi.Error{Code: 404}
	err := retryWithBackoff(context.Background(), func() error {
		calls++
		return wantErr
	}, 5, time.Millisecond, 5*time.Millisecond)

	if err != wantErr {
		t.Errorf("expected the original 404 error to be returned unwrapped, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly 1 call (no retry on non-rate-limit error), got %d", calls)
	}
}

func TestRetryWithBackoffGivesUpAfterMaxRetries(t *testing.T) {
	calls := 0
	err := retryWithBackoff(context.Background(), func() error {
		calls++
		return &googleapi.Error{Code: 429}
	}, 3, time.Millisecond, time.Millisecond)

	if err == nil {
		t.Fatal("expected an error after exhausting retries, got nil")
	}
	if calls != 3 {
		t.Errorf("expected exactly 3 calls (maxRetries), got %d", calls)
	}
}

func TestRetryWithBackoffRespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	err := retryWithBackoff(ctx, func() error {
		calls++
		return &googleapi.Error{Code: 429}
	}, 5, 50*time.Millisecond, 100*time.Millisecond)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly 1 call before cancellation was observed, got %d", calls)
	}
}
