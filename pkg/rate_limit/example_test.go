package ratelimit_test

import (
	"time"

	ratelimit "github.com/kydenul/template-go/pkg/rate_limit"
	"github.com/redis/go-redis/v9"
)

// 示例：使用默认配置创建限流器
func ExampleNewRateLimiter_default() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 使用默认配置
	limiter := ratelimit.NewRateLimiter(rdb)
	_ = limiter
}

// 示例：使用自定义配置创建限流器
func ExampleNewRateLimiter_withOptions() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 使用 Option 模式自定义配置
	limiter := ratelimit.NewRateLimiter(
		rdb,
		ratelimit.WithRedisTimeout(200*time.Millisecond),
		ratelimit.WithGlobalQPS(2000),
		ratelimit.WithIPRatePerMinute(200),
		ratelimit.WithUserRatePerMinute(120),
		ratelimit.WithFallbackAllow(false), // Redis 故障时拒绝请求
	)
	_ = limiter
}

// 示例：只自定义部分配置，其余使用默认值
func ExampleNewRateLimiter_partial() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 只修改部分配置，其余保持默认
	limiter := ratelimit.NewRateLimiter(
		rdb,
		ratelimit.WithGlobalQPS(5000),
		ratelimit.WithRedisTimeout(150*time.Millisecond),
	)
	_ = limiter
}
