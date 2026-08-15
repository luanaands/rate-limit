package middleware

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func TestRequestToken(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		want    string
		wantErr bool
	}{
		{
			name:   "api key",
			apiKey: "token-123",
			want:   "token-123",
		},
		{
			name: "empty api key",
		},
		{
			name:    "nil request",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r *http.Request
			if !tt.wantErr {
				r = &http.Request{Header: make(http.Header)}
				r.Header.Set("API_KEY", tt.apiKey)
			}

			got, err := requestToken(r)
			if (err != nil) != tt.wantErr {
				t.Fatalf("requestToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("requestToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRateLimiterConfig(t *testing.T) {
	blockDuration := 5 * time.Minute
	ratelimiter := NewRateLimiterConfig(100, 10, blockDuration, nil)

	if ratelimiter.TokenLimit != 100 {
		t.Fatalf("TokenLimit = %d, want 100", ratelimiter.TokenLimit)
	}
	if ratelimiter.IPLimit != 10 {
		t.Fatalf("IPLimit = %d, want 10", ratelimiter.IPLimit)
	}
	if ratelimiter.BlockDuration != blockDuration {
		t.Fatalf("BlockDuration = %s, want %s", ratelimiter.BlockDuration, blockDuration)
	}
	if ratelimiter.RedisClient != nil {
		t.Fatal("expected RedisClient to be nil")
	}
}

func TestRateLimiterBlocksUntilConfiguredDurationExpires(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	ratelimiter := NewRateLimiterConfig(1, 1, time.Minute, client)
	request := requestWithToken("token-123")

	allowed, err := ratelimiter.AllowByToken(context.Background(), request)
	if err != nil {
		t.Fatalf("first request returned an error: %v", err)
	}
	if !allowed {
		t.Fatal("expected first request to be allowed")
	}

	allowed, err = ratelimiter.AllowByToken(context.Background(), request)
	if err != nil {
		t.Fatalf("request that exceeds the limit returned an error: %v", err)
	}
	if allowed {
		t.Fatal("expected request that exceeds the limit to be blocked")
	}

	server.FastForward(30 * time.Second)
	allowed, err = ratelimiter.AllowByToken(context.Background(), request)
	if err != nil {
		t.Fatalf("blocked request returned an error: %v", err)
	}
	if allowed {
		t.Fatal("expected request to remain blocked before the duration expires")
	}

	server.FastForward(31 * time.Second)
	allowed, err = ratelimiter.AllowByToken(context.Background(), request)
	if err != nil {
		t.Fatalf("request after the block expired returned an error: %v", err)
	}
	if !allowed {
		t.Fatal("expected request to be allowed after the block expired")
	}
}

func TestRateLimiterUsesIndependentBlockKeys(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	ratelimiter := NewRateLimiterConfig(1, 1, time.Minute, client)

	tokenRequest := requestWithToken("token-123")
	ipRequest := &http.Request{
		RemoteAddr: "192.0.2.1:8080",
		Header:     make(http.Header),
	}

	if allowed, err := ratelimiter.AllowByToken(context.Background(), tokenRequest); err != nil || !allowed {
		t.Fatalf("first token request: allowed=%v, error=%v", allowed, err)
	}
	if allowed, err := ratelimiter.AllowByToken(context.Background(), tokenRequest); err != nil || allowed {
		t.Fatalf("second token request: allowed=%v, error=%v", allowed, err)
	}
	if allowed, err := ratelimiter.AllowByIP(context.Background(), ipRequest); err != nil || !allowed {
		t.Fatalf("first IP request: allowed=%v, error=%v", allowed, err)
	}
}

func requestWithToken(token string) *http.Request {
	request := &http.Request{Header: make(http.Header)}
	request.Header.Set("API_KEY", token)
	return request
}

func TestAllowByTokenRequiresAPIKey(t *testing.T) {
	ratelimiter := &RateLimiterRedis{}
	request := &http.Request{Header: make(http.Header)}

	_, err := ratelimiter.AllowByToken(context.Background(), request)

	if err == nil {
		t.Fatal("expected an error when API_KEY is empty")
	}
}

func TestAllowSelectsTokenOrIP(t *testing.T) {
	ratelimiter := &RateLimiterRedis{}

	tests := []struct {
		name       string
		apiKey     string
		remoteAddr string
		wantError  string
	}{
		{
			name:       "token",
			apiKey:     "token-123",
			remoteAddr: "invalid",
			wantError:  "redis client is nil",
		},
		{
			name:       "ip",
			remoteAddr: "invalid",
			wantError:  "failed to get client IP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &http.Request{RemoteAddr: tt.remoteAddr, Header: make(http.Header)}
			request.Header.Set("API_KEY", tt.apiKey)

			_, err := ratelimiter.Allow(context.Background(), request)
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("Allow() error = %v, want it to contain %q", err, tt.wantError)
			}
		})
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		wantIP     string
		wantErr    bool
	}{
		{
			name:       "ipv4 with port",
			remoteAddr: "192.0.2.1:8080",
			wantIP:     "192.0.2.1",
		},
		{
			name:       "ipv6 with port",
			remoteAddr: "[2001:db8::1]:8080",
			wantIP:     "2001:db8::1",
		},
		{
			name:       "ip without port",
			remoteAddr: "192.0.2.1",
			wantIP:     "192.0.2.1",
		},
		{
			name:       "invalid address",
			remoteAddr: "client",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := clientIP(&http.Request{RemoteAddr: tt.remoteAddr})
			if (err != nil) != tt.wantErr {
				t.Fatalf("clientIP() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantIP {
				t.Fatalf("clientIP() = %q, want %q", got, tt.wantIP)
			}
		})
	}
}

func TestClientIPNilRequest(t *testing.T) {
	if _, err := clientIP(nil); err == nil {
		t.Fatal("expected an error for a nil request")
	}
}
