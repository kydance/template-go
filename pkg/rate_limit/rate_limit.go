package ratelimit

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

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

	// 策略启用配置
	enableGlobalRateLimit bool // 是否启用全局限流
	enableIPRateLimit     bool // 是否启用 IP 限流
	enableUserRateLimit   bool // 是否启用用户限流

	enableHealthCheck bool // 是否启用 Redis 健康检查
	enableMetrics     bool // 是否启用 Metrics 监控
}

// defaultConfig 返回默认配置
func defaultConfig() *config {
	return &config{
		redisTimeout:          100 * time.Millisecond,
		fallbackAllow:         true,
		fallbackRemaining:     999,
		fallbackRetryAfter:    time.Second,
		globalQPS:             1000,
		ipRatePerMinute:       100,
		userRatePerMinute:     60,
		enableGlobalRateLimit: true, // 默认全部启用
		enableIPRateLimit:     true,
		enableUserRateLimit:   true,
		enableHealthCheck:     true,  // 默认启用健康检查
		enableMetrics:         false, // 默认不启用监控（避免性能开销）
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

func WithEnableGlobalRateLimit(enable bool) Option {
	return func(c *config) { c.enableGlobalRateLimit = enable }
}

func WithEnableIPRateLimit(enable bool) Option {
	return func(c *config) { c.enableIPRateLimit = enable }
}

func WithEnableUserRateLimit(enable bool) Option {
	return func(c *config) { c.enableUserRateLimit = enable }
}

func WithEnableHealthCheck(enable bool) Option {
	return func(c *config) { c.enableHealthCheck = enable }
}

func WithEnableMetrics(enable bool) Option {
	return func(c *config) { c.enableMetrics = enable }
}

type RateLimiter struct {
	limiter *Limiter
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
		limiter: NewLimiter(rdb),
		rdb:     rdb,
		config:  cfg,
	}
}

func (r *RateLimiter) AllowWithFallback(
	ctx context.Context,
	key string,
	limit Limit,
) (*Result, error) {
	// Set timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, r.config.redisTimeout)
	defer cancel()

	// Check health if enabled
	if r.config.enableHealthCheck {
		if err := r.healthCheck(ctxWithTimeout); err != nil {
			log.Printf("Redis check health fail: %v, use fallback", err)
			return r.fallback(), nil
		}
	}

	res, err := r.limiter.Allow(ctxWithTimeout, key, limit)
	if err != nil {
		log.Printf("Redis rate limit fail: %v, use fallback", err)
		return r.fallback(), nil
	}

	return res, nil
}

// healthCheck check redis health
func (r *RateLimiter) healthCheck(ctx context.Context) error {
	err := r.rdb.Ping(ctx).Err()
	if err != nil {
		log.Printf("Redis health check fail: %v", err)
		return err
	}

	log.Println("Redis health check success")
	return nil
}

// fallback
func (r *RateLimiter) fallback() *Result {
	// Fallback => Allow
	if r.config.fallbackAllow {
		return &Result{
			Allowed:   1,
			Remaining: r.config.fallbackRemaining, // Fake Value -> Fallback
		}
	}

	// Fallback => Deny
	return &Result{
		Allowed:    0,
		Remaining:  0,
		RetryAfter: r.config.fallbackRetryAfter,
	}
}

func (r *RateLimiter) CheckMultiDimensional(
	ctx context.Context,
	userID, userIP, endpoint string,
) (bool, string) {
	// Global rate limit => the QPS of all cluster
	if r.config.enableGlobalRateLimit {
		globalKey := "global:ratelimit:" + endpoint
		res, err := r.AllowWithFallback(ctx, globalKey, PerSecond(r.config.globalQPS))

		if err != nil && r.config.enableMetrics {
			metrics.redisErrors.Add(1)
		}

		isFallback := res.Remaining == r.config.fallbackRemaining
		allowed := res.Allowed > 0
		r.recordMetrics("global", allowed, isFallback)

		if err != nil || !allowed {
			return false, "GlobalRateLimit"
		}
	}

	// IP Rate Limit => the QPS of single IP
	if r.config.enableIPRateLimit {
		ipKey := "ratelimit:ip:" + userIP
		res, err := r.AllowWithFallback(ctx, ipKey, PerMinute(r.config.ipRatePerMinute))

		if err != nil && r.config.enableMetrics {
			metrics.redisErrors.Add(1)
		}

		isFallback := res.Remaining == r.config.fallbackRemaining
		allowed := res.Allowed > 0
		r.recordMetrics("ip", allowed, isFallback)

		if err != nil || !allowed {
			return false, "IPRateLimit"
		}
	}

	// User rate limit => the QPS of single user
	if r.config.enableUserRateLimit && userID != "" {
		userKey := "ratelimit:user:" + userID + ":" + endpoint
		res, err := r.AllowWithFallback(
			ctx,
			userKey,
			PerMinute(r.config.userRatePerMinute),
		)

		if err != nil && r.config.enableMetrics {
			metrics.redisErrors.Add(1)
		}

		isFallback := res.Remaining == r.config.fallbackRemaining
		allowed := res.Allowed > 0
		r.recordMetrics("user", allowed, isFallback)

		if err != nil || !allowed {
			return false, "UserRateLimit"
		}
	}

	return true, ""
}

type DimensionMetrics struct {
	totalRequests    atomic.Int64
	allowedRequests  atomic.Int64
	rejectedRequests atomic.Int64
}

type RateLimitMetrics struct {
	global            DimensionMetrics
	ip                DimensionMetrics
	user              DimensionMetrics
	redisErrors       atomic.Int64
	fallbackActivated atomic.Int64
}

var metrics RateLimitMetrics

// recordMetrics 记录维度指标
func (r *RateLimiter) recordMetrics(dimension string, allowed, isFallback bool) {
	if !r.config.enableMetrics {
		return
	}

	var dim *DimensionMetrics
	switch dimension {
	case "global":
		dim = &metrics.global
	case "ip":
		dim = &metrics.ip
	case "user":
		dim = &metrics.user
	default:
		return
	}

	dim.totalRequests.Add(1)
	if allowed {
		dim.allowedRequests.Add(1)
	} else {
		dim.rejectedRequests.Add(1)
	}

	if isFallback {
		metrics.fallbackActivated.Add(1)
	}
}

// MetricsHandler export Prometheus metrics
func MetricsHandler(w http.ResponseWriter, _ *http.Request) {
	// Global dimension metrics
	_, _ = fmt.Fprintf(w, "# HELP ratelimit_total_requests Total rate limit checks by dimension\n")
	_, _ = fmt.Fprintf(w, "# TYPE ratelimit_total_requests counter\n")
	_, _ = fmt.Fprintf(w, "ratelimit_total_requests{dimension=\"global\"} %d\n",
		metrics.global.totalRequests.Load())
	_, _ = fmt.Fprintf(w, "ratelimit_total_requests{dimension=\"ip\"} %d\n",
		metrics.ip.totalRequests.Load())
	_, _ = fmt.Fprintf(w, "ratelimit_total_requests{dimension=\"user\"} %d\n",
		metrics.user.totalRequests.Load())

	_, _ = fmt.Fprintf(w, "# HELP ratelimit_allowed_requests Allowed requests by dimension\n")
	_, _ = fmt.Fprintf(w, "# TYPE ratelimit_allowed_requests counter\n")
	_, _ = fmt.Fprintf(w, "ratelimit_allowed_requests{dimension=\"global\"} %d\n",
		metrics.global.allowedRequests.Load())
	_, _ = fmt.Fprintf(
		w,
		"ratelimit_allowed_requests{dimension=\"ip\"} %d\n",
		metrics.ip.allowedRequests.Load(),
	)
	_, _ = fmt.Fprintf(
		w,
		"ratelimit_allowed_requests{dimension=\"user\"} %d\n",
		metrics.user.allowedRequests.Load(),
	)

	_, _ = fmt.Fprintf(
		w,
		"# HELP ratelimit_rejected_requests Rejected requests by dimension\n",
	)
	_, _ = fmt.Fprintf(w, "# TYPE ratelimit_rejected_requests counter\n")
	_, _ = fmt.Fprintf(
		w,
		"ratelimit_rejected_requests{dimension=\"global\"} %d\n",
		metrics.global.rejectedRequests.Load(),
	)
	_, _ = fmt.Fprintf(
		w,
		"ratelimit_rejected_requests{dimension=\"ip\"} %d\n",
		metrics.ip.rejectedRequests.Load(),
	)
	_, _ = fmt.Fprintf(
		w,
		"ratelimit_rejected_requests{dimension=\"user\"} %d\n",
		metrics.user.rejectedRequests.Load(),
	)

	// Redis errors (global counter)
	_, _ = fmt.Fprintf(w, "# HELP ratelimit_redis_errors Redis errors\n")
	_, _ = fmt.Fprintf(w, "# TYPE ratelimit_redis_errors counter\n")
	_, _ = fmt.Fprintf(w, "ratelimit_redis_errors %d\n", metrics.redisErrors.Load())

	// Fallback activations (global counter)
	_, _ = fmt.Fprintf(w, "# HELP ratelimit_fallback_activated Fallback activations\n")
	_, _ = fmt.Fprintf(w, "# TYPE ratelimit_fallback_activated counter\n")
	_, _ = fmt.Fprintf(
		w,
		"ratelimit_fallback_activated %d\n",
		metrics.fallbackActivated.Load(),
	)
}
