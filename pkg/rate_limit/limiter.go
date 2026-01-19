// Package ratelimit provides a Redis-based rate limiter using GCRA algorithm.
package ratelimit

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisPrefix = "rate:"

// Limit defines the rate limit parameters.
type Limit struct {
	Rate   int           // Number of events allowed per period
	Burst  int           // Maximum burst size (bucket capacity)
	Period time.Duration // Time period for the rate
}

// PerSecond creates a Limit that allows rate events per second with burst equal to rate.
func PerSecond(rate int) Limit {
	return Limit{
		Rate:   rate,
		Burst:  rate,
		Period: time.Second,
	}
}

// PerMinute creates a Limit that allows rate events per minute with burst equal to rate.
func PerMinute(rate int) Limit {
	return Limit{
		Rate:   rate,
		Burst:  rate,
		Period: time.Minute,
	}
}

// PerHour creates a Limit that allows rate events per hour with burst equal to rate.
func PerHour(rate int) Limit {
	return Limit{
		Rate:   rate,
		Burst:  rate,
		Period: time.Hour,
	}
}

// Result represents the result of a rate limit check.
type Result struct {
	// Allowed is the number of events that are allowed.
	// It's either the requested n or 0 if the limit is exceeded.
	Allowed int

	// Remaining is the number of remaining events in the current window.
	Remaining int

	// RetryAfter is the time after which the request can be retried.
	// -1 means no retry is needed (request was allowed).
	RetryAfter time.Duration

	// ResetAfter is the time after which the rate limiter will be reset.
	ResetAfter time.Duration
}

// Limiter is a Redis-based rate limiter using GCRA algorithm.
type Limiter struct {
	rdb redis.UniversalClient
}

// NewLimiter creates a new rate limiter with the given Redis client.
func NewLimiter(rdb redis.UniversalClient) *Limiter {
	return &Limiter{rdb: rdb}
}

// Allow is a shorthand for AllowN(ctx, key, limit, 1).
func (l *Limiter) Allow(ctx context.Context, key string, limit Limit) (*Result, error) {
	return l.AllowN(ctx, key, limit, 1)
}

// AllowN checks if n events can occur for the given key.
func (l *Limiter) AllowN(ctx context.Context, key string, limit Limit, n int) (*Result, error) {
	now := time.Now()
	nowSec := float64(now.Unix())
	nowMicrosec := float64(now.Nanosecond()) / 1000.0

	values := []any{
		limit.Burst,
		limit.Rate,
		limit.Period.Seconds(),
		n,
		nowSec,
		nowMicrosec,
	}

	v, err := allowN.Run(ctx, l.rdb, []string{redisPrefix + key}, values...).Result()
	if err != nil {
		return nil, err
	}

	return parseResult(v)
}

// AllowAtMost checks if at most n events can occur for the given key.
// It returns the number of allowed events which may be less than n.
func (l *Limiter) AllowAtMost(
	ctx context.Context,
	key string,
	limit Limit,
	n int,
) (*Result, error) {
	now := time.Now()
	nowSec := float64(now.Unix())
	nowMicrosec := float64(now.Nanosecond()) / 1000.0

	values := []any{
		limit.Burst,
		limit.Rate,
		limit.Period.Seconds(),
		n,
		nowSec,
		nowMicrosec,
	}

	v, err := allowAtMost.Run(ctx, l.rdb, []string{redisPrefix + key}, values...).Result()
	if err != nil {
		return nil, err
	}

	return parseResult(v)
}

// Reset resets the rate limiter for the given key.
func (l *Limiter) Reset(ctx context.Context, key string) error {
	return l.rdb.Del(ctx, redisPrefix+key).Err()
}

func parseResult(v any) (*Result, error) {
	values, _ := v.([]any)

	val1, _ := values[2].(string)
	retryAfter, err := strconv.ParseFloat(val1, 64)
	if err != nil {
		return nil, err
	}

	val2, _ := values[3].(string)
	resetAfter, err := strconv.ParseFloat(val2, 64)
	if err != nil {
		return nil, err
	}

	allowed, _ := values[0].(int64)
	remaining, _ := values[1].(int64)

	return &Result{
		Allowed:    int(allowed),
		Remaining:  int(remaining),
		RetryAfter: dur(retryAfter),
		ResetAfter: dur(resetAfter),
	}, nil
}

func dur(f float64) time.Duration {
	if f < 0 {
		return -1
	}
	return time.Duration(f * float64(time.Second))
}
