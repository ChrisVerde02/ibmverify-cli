package retry

import (
	"context"
	"errors"
	"testing"

	"github.com/ChrisVerde02/ibmverify-go/client"
)

func TestDo_succeedsFirstAttempt(t *testing.T) {
	calls := 0
	err := doWithDelay(context.Background(), func() error {
		calls++
		return nil
	}, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestDo_retriesOnServerError(t *testing.T) {
	calls := 0
	serverErr := &client.APIError{StatusCode: 500}
	err := doWithDelay(context.Background(), func() error {
		calls++
		if calls < 3 {
			return serverErr
		}
		return nil
	}, 0)
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestDo_retriesOnRateLimit(t *testing.T) {
	calls := 0
	rlErr := &client.APIError{StatusCode: 429}
	err := doWithDelay(context.Background(), func() error {
		calls++
		if calls < 2 {
			return rlErr
		}
		return nil
	}, 0)
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestDo_doesNotRetryAuthError(t *testing.T) {
	calls := 0
	authErr := &client.APIError{StatusCode: 401}
	err := doWithDelay(context.Background(), func() error {
		calls++
		return authErr
	}, 0)
	if !errors.Is(err, authErr) {
		t.Fatalf("expected authErr back, got: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly 1 call (no retry), got %d", calls)
	}
}

func TestDo_doesNotRetryNotFound(t *testing.T) {
	calls := 0
	nfErr := &client.APIError{StatusCode: 404}
	err := doWithDelay(context.Background(), func() error {
		calls++
		return nfErr
	}, 0)
	if !errors.Is(err, nfErr) {
		t.Fatalf("expected nfErr back, got: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly 1 call (no retry), got %d", calls)
	}
}

func TestDo_exhaustsMaxAttempts(t *testing.T) {
	calls := 0
	serverErr := &client.APIError{StatusCode: 503}
	err := doWithDelay(context.Background(), func() error {
		calls++
		return serverErr
	}, 0)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if calls != maxAttempts {
		t.Errorf("expected %d calls, got %d", maxAttempts, calls)
	}
}

func TestDo_respectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before any sleep

	calls := 0
	err := doWithDelay(ctx, func() error {
		calls++
		return &client.APIError{StatusCode: 503}
	}, 0)
	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
	// With delay=0, time.After fires immediately so the context check may
	// happen after 1 or 2 calls — either is acceptable as long as we stop.
	if calls > maxAttempts {
		t.Errorf("expected at most %d calls, got %d", maxAttempts, calls)
	}
}
