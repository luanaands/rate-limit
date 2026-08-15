package middleware

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

type fakeRateLimiter struct {
	allowed bool
	request *http.Request
	err     error
}

func (f *fakeRateLimiter) Allow(_ context.Context, r *http.Request) (bool, error) {
	f.request = r
	return f.allowed, f.err
}

func TestRateLimitProcessorDelegatesToStrategy(t *testing.T) {
	strategy := &fakeRateLimiter{allowed: true}
	processor := NewRateLimitProcessor(strategy)
	request := &http.Request{}

	allowed, err := processor.Allow(context.Background(), request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected request to be allowed")
	}
	if strategy.request != request {
		t.Fatal("expected processor to pass the request to the strategy")
	}
}

func TestRateLimitProcessorPropagatesStrategyError(t *testing.T) {
	wantErr := errors.New("strategy failed")
	processor := NewRateLimitProcessor(&fakeRateLimiter{err: wantErr})

	_, err := processor.Allow(context.Background(), &http.Request{})

	if !errors.Is(err, wantErr) {
		t.Fatalf("got error %v, want %v", err, wantErr)
	}
}

func TestRateLimitProcessorRejectsNilStrategy(t *testing.T) {
	processor := NewRateLimitProcessor(nil)

	allowed, err := processor.Allow(context.Background(), &http.Request{})

	if allowed {
		t.Fatal("expected request to be rejected")
	}
	if err == nil {
		t.Fatal("expected an error")
	}
}
