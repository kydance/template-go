package ratelimit

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type config struct {
	// Redis Timeout config
	redisTimeout time.Duration

	// 降级配置
	fallbackAllow      bool // Redis 故障时是否放行
	fallbackRemaining  int  // 降级时的剩余配额标识值
	fallbackRetryAfter time.Duration

	// 全局限流配置, 全局每秒请求数
	globalQPS int

	// IP 限流配置, 单个 IP 每分钟请求数
	ipRatePerMinute int

	// 用户限流配置, 单个用户每分钟请求数
	userRatePerMinute int
}

// defaultConfig 返回默认配置
func defaultConfig() *config {
	return &config{
		redisTimeout:       100 * time.Millisecond,
		fallbackAllow:      true,
		fallbackRemaining:  999,
		fallbackRetryAfter: time.Second,
		globalQPS:          1000,
		ipRatePerMinute:    100,
		userRatePerMinute:  60,
	}
}

type Option func(*config)

func WithRedisTimeout(timeout time.Duration) Option {
	return func(c *config) { c.redisTimeout = timeout }
}

func WithFallbackAllow(allow bool) Option {
	return func(c *config) { c.fallbackAllow = allow }
}

func WithFallbackRemaining(remaining int) Option {
	return func(c *config) { c.fallbackRemaining = remaining }
}

func WithFallbackRetryAfter(retryAfter time.Duration) Option {
	return func(c *config) { c.fallbackRetryAfter = retryAfter }
}

func WithGlobalQPS(qps int) Option {
	return func(c *config) { c.globalQPS = qps }
}

func WithIPRatePerMinute(rate int) Option {
	return func(c *config) { c.ipRatePerMinute = rate }
}

func WithUserRatePerMinute(rate int) Option {
	return func(c *config) { c.userRatePerMinute = rate }
}

type RateLimiter struct {
	limiter *redis_rate.Limiter
	rdb     redis.UniversalClient
	config  *config
}

// NewRateLimiter 创建限流器实例，支持可选配置
func NewRateLimiter(rdb redis.UniversalClient, opts ...Option) *RateLimiter {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	return &RateLimiter{
		limiter: redis_rate.NewLimiter(rdb),
		rdb:     rdb,
		config:  cfg,
	}
}

func (r *RateLimiter) AllowWithFallback(
	ctx context.Context,
	key string,
	limit redis_rate.Limit,
) (*redis_rate.Result, error) {
	// Set timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, r.config.redisTimeout)
	defer cancel()

	// Check health
	if err := r.healthCheck(ctxWithTimeout); err != nil {
		log.Printf("Redis check health fail: %v, use fallback", err)
		return r.fallback(), nil
	}

	//
	res, err := r.limiter.Allow(ctxWithTimeout, key, limit)
	if err != nil {
		log.Printf("Redis rate limit fail: %v, use fallback", err)
		return r.fallback(), nil
	}

	return res, nil
}

// healthCheck check redis health
func (r *RateLimiter) healthCheck(ctx context.Context) error { return r.rdb.Ping(ctx).Err() }

// fallback
func (r *RateLimiter) fallback() *redis_rate.Result {
	// Fallback => Allow
	if r.config.fallbackAllow {
		return &redis_rate.Result{
			Allowed:   1,
			Remaining: r.config.fallbackRemaining, // Fake Value -> Fallback
		}
	}

	// Fallback => Deny
	return &redis_rate.Result{
		Allowed:    0,
		Remaining:  0,
		RetryAfter: r.config.fallbackRetryAfter,
	}
}

func (r *RateLimiter) CheckMultiDimensional(
	ctx context.Context,
	userID, userIP, endpoint string,
) (bool, string) {
	// Golbal rate limit => the QPS of all cluster
	globalKey := "global:ratelimit:" + endpoint
	res, err := r.AllowWithFallback(ctx, globalKey, redis_rate.PerSecond(r.config.globalQPS))
	if err != nil || res.Allowed == 0 {
		return false, "GlobalRateLimt"
	}

	// IP Rate Limit => the QPS of single IP
	ipKey := "ratelimit:ip:" + userIP
	res, err = r.AllowWithFallback(ctx, ipKey, redis_rate.PerMinute(r.config.ipRatePerMinute))
	if err != nil || res.Allowed == 0 {
		return false, "IPRateLimit"
	}

	// User rate limit => the QPS of single user
	if userID != "" {
		userKey := "ratelimit:user:" + userID + ":" + endpoint
		res, err = r.AllowWithFallback(
			ctx,
			userKey,
			redis_rate.PerMinute(r.config.userRatePerMinute),
		)
		if err != nil || res.Allowed == 0 {
			return false, "UserRateLimit"
		}
	}

	return true, ""
}

type RateLimitMetrics struct {
	totalRequests     atomic.Int64
	allowedRequests   atomic.Int64
	rejectedRequests  atomic.Int64
	redisErrors       atomic.Int64
	fallbackActivated atomic.Int64
}

var metrics RateLimitMetrics

func (r *RateLimiter) CheckWithMetrics(
	ctx context.Context,
	key string,
	limit redis_rate.Limit,
) (*redis_rate.Result, error) {
	metrics.totalRequests.Add(1)

	res, err := r.AllowWithFallback(ctx, key, limit)
	if err != nil {
		metrics.redisErrors.Add(1)
	}

	if res.Allowed > 0 {
		metrics.allowedRequests.Add(1)
	} else {
		metrics.rejectedRequests.Add(1)
	}

	// Check if fallback is activated
	if res.Remaining == r.config.fallbackRemaining {
		metrics.fallbackActivated.Add(1)
	}

	return res, err
}

// MetricsHandler export Prometheus metrics
func MetricsHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, "# HELP ratelimit_total_requests Total rate limit checks\n")
	_, _ = fmt.Fprintf(w, "# TYPE ratelimit_total_requests counter\n")
	_, _ = fmt.Fprintf(w, "ratelimit_total_requests %d\n", metrics.totalRequests.Load())

	_, _ = fmt.Fprintf(w, "# HELP ratelimit_allowed_requests Allowed requests\n")
	_, _ = fmt.Fprintf(w, "# TYPE ratelimit_allowed_requests counter\n")
	_, _ = fmt.Fprintf(w, "ratelimit_allowed_requests %d\n", metrics.allowedRequests.Load())

	_, _ = fmt.Fprintf(w, "# HELP ratelimit_rejected_requests Rejected requests\n")
	_, _ = fmt.Fprintf(w, "# TYPE ratelimit_rejected_requests counter\n")
	_, _ = fmt.Fprintf(w, "ratelimit_rejected_requests %d\n", metrics.rejectedRequests.Load())

	_, _ = fmt.Fprintf(w, "# HELP ratelimit_redis_errors Redis errors\n")
	_, _ = fmt.Fprintf(w, "# TYPE ratelimit_redis_errors counter\n")
	_, _ = fmt.Fprintf(w, "ratelimit_redis_errors %d\n", metrics.redisErrors.Load())

	_, _ = fmt.Fprintf(w, "# HELP ratelimit_fallback_activated Fallback activations\n")
	_, _ = fmt.Fprintf(w, "# TYPE ratelimit_fallback_activated counter\n")
	_, _ = fmt.Fprintf(w, "ratelimit_fallback_activated %d\n", metrics.fallbackActivated.Load())
}
