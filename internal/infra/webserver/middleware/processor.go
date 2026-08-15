package middleware

import (
	"context"
	"errors"
	"net/http"
)

type RateLimitProcessor struct {
	strategy RateLimiterInterface
}

var _ RateLimiterInterface = (*RateLimitProcessor)(nil)

func NewRateLimitProcessor(strategy RateLimiterInterface) *RateLimitProcessor {
	return &RateLimitProcessor{strategy: strategy}
}

func (p *RateLimitProcessor) Allow(ctx context.Context, r *http.Request) (bool, error) {
	if p == nil || p.strategy == nil {
		return false, errors.New("rate limiter strategy is nil")
	}

	return p.strategy.Allow(ctx, r)
}
