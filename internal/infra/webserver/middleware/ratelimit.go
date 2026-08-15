package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
)

const rateLimitWindow = time.Second
const blockedKeyPrefix = "blocked:"

type RateLimiterRedis struct {
	TokenLimit    int
	IPLimit       int
	BlockDuration time.Duration
	RedisClient   *redis.Client
	sequence      uint64
}

func NewRateLimiterConfig(tokenLimit, ipLimit int, blockDuration time.Duration, rdb *redis.Client) *RateLimiterRedis {
	return &RateLimiterRedis{
		TokenLimit:    tokenLimit,
		IPLimit:       ipLimit,
		BlockDuration: blockDuration,
		RedisClient:   rdb,
	}
}

// Allow checks if a request is allowed based on the sliding window algorithm
func (rlc *RateLimiterRedis) Allow(ctx context.Context, r *http.Request) (bool, error) {
	token, err := requestToken(r)
	if err == nil && token != "" {
		return rlc.AllowByToken(ctx, r)
	}
	return rlc.AllowByIP(ctx, r)
}

func (rlc *RateLimiterRedis) AllowByToken(ctx context.Context, r *http.Request) (bool, error) {
	token, err := requestToken(r)
	if err != nil {
		return false, fmt.Errorf("failed to get API key: %w", err)
	}
	if token == "" {
		return false, fmt.Errorf("API_KEY header is empty")
	}

	return rlc.allow(ctx, "token:"+token, rlc.TokenLimit)
}

func (rlc *RateLimiterRedis) AllowByIP(ctx context.Context, r *http.Request) (bool, error) {
	ip, err := clientIP(r)
	if err != nil {
		return false, fmt.Errorf("failed to get client IP: %w", err)
	}

	return rlc.allow(ctx, "ip:"+ip, rlc.IPLimit)
}

func (rlc *RateLimiterRedis) allow(ctx context.Context, key string, limit int) (bool, error) {
	if rlc.RedisClient == nil {
		return false, fmt.Errorf("redis client is nil")
	}

	blockedKey := blockedKeyPrefix + key
	blocked, err := rlc.RedisClient.Exists(ctx, blockedKey).Result()
	if err != nil {
		return false, fmt.Errorf("redis transaction failed: %w", err)
	}
	if blocked > 0 {
		return false, nil
	}

	now := time.Now().Unix()
	windowSeconds := int64(rateLimitWindow / time.Second)

	// Use a Redis transaction (MULTI/EXEC) so concurrent requests are counted atomically.
	pipe := rlc.RedisClient.TxPipeline()

	minScore := now - windowSeconds
	pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(minScore, 10))

	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d-%d", now, atomic.AddUint64(&rlc.sequence, 1)),
	})

	countCmd := pipe.ZCount(ctx, key, strconv.FormatInt(minScore, 10), "+inf")

	pipe.Expire(ctx, key, rateLimitWindow+time.Second) // Add a buffer for safety

	// Execute all commands in the pipeline atomically
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("redis transaction failed: %w", err)
	}

	// Get the count from the executed command
	currentRequests, err := countCmd.Result()
	if err != nil {
		return false, fmt.Errorf("failed to get request count from redis: %w", err)
	}

	if int(currentRequests) <= limit {
		return true, nil
	}

	if rlc.BlockDuration > 0 {
		set, _ := rlc.RedisClient.SetNX(ctx, blockedKey, "1", rlc.BlockDuration).Result()
		if !set {
			return false, nil
		}
	}

	return false, nil
}

func clientIP(r *http.Request) (string, error) {
	if r == nil {
		return "", fmt.Errorf("request is nil")
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && net.ParseIP(host) != nil {
		return host, nil
	}
	if net.ParseIP(r.RemoteAddr) != nil {
		return r.RemoteAddr, nil
	}

	return "", fmt.Errorf("invalid remote address %q", r.RemoteAddr)
}

func requestToken(r *http.Request) (string, error) {
	if r == nil {
		return "", fmt.Errorf("request is nil")
	}
	return strings.TrimSpace(r.Header.Get("API_KEY")), nil
}
