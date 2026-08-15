package middleware

import (
	"context"
	"net/http"
)

type RateLimiterInterface interface {
	Allow(ctx context.Context, r *http.Request) (bool, error)
}
