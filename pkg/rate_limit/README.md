# Rate Limiter

基于 Redis 的高性能分布式限流器，支持多维度限流和降级策略。

## 特性

- **多维度限流**：支持全局、IP、用户三个维度的限流
- **降级策略**：Redis 故障时支持配置降级行为（放行/拒绝）
- **健康检查**：内置 Redis 健康检查机制
- **指标监控**：内置 Prometheus 指标导出
- **Option 模式**：灵活的配置方式，支持按需自定义

## 安装

```bash
go get github.com/kydenul/template-go/pkg/rate_limit
```

## 快速开始

### 使用默认配置

```go
import (
    ratelimit "github.com/kydenul/template-go/pkg/rate_limit"
    "github.com/redis/go-redis/v9"
)

// 创建 Redis 客户端
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// 使用默认配置创建限流器
limiter := ratelimit.NewRateLimiter(rdb)
```

### 使用自定义配置

```go
// 使用 Option 模式自定义配置
limiter := ratelimit.NewRateLimiter(
    rdb,
    ratelimit.WithRedisTimeout(200*time.Millisecond),
    ratelimit.WithGlobalQPS(2000),
    ratelimit.WithIPRatePerMinute(200),
    ratelimit.WithUserRatePerMinute(120),
    ratelimit.WithFallbackAllow(false),
)
```

### 只自定义部分配置

```go
// 只修改需要的配置，其余保持默认值
limiter := ratelimit.NewRateLimiter(
    rdb,
    ratelimit.WithGlobalQPS(5000),
    ratelimit.WithRedisTimeout(150*time.Millisecond),
)
```

## 配置选项

| Option 函数 | 说明 | 默认值 |
|------------|------|--------|
| `WithRedisTimeout(duration)` | Redis 操作超时时间 | 100ms |
| `WithFallbackAllow(bool)` | Redis 故障时是否放行 | true |
| `WithFallbackRemaining(int)` | 降级时的剩余配额标识值 | 999 |
| `WithFallbackRetryAfter(duration)` | 降级拒绝时的重试等待时间 | 1s |
| `WithGlobalQPS(int)` | 全局每秒请求数限制 | 1000 |
| `WithIPRatePerMinute(int)` | 单个 IP 每分钟请求数限制 | 100 |
| `WithUserRatePerMinute(int)` | 单个用户每分钟请求数限制 | 60 |

## 使用示例

### 基本限流

```go
ctx := context.Background()
key := "api:user:123"
limit := redis_rate.PerSecond(10) // 每秒 10 次

res, err := limiter.AllowWithFallback(ctx, key, limit)
if err != nil || res.Allowed == 0 {
    // 请求被限流
    log.Printf("Rate limited. Retry after: %v", res.RetryAfter)
    return
}

// 请求通过
log.Printf("Request allowed. Remaining: %d", res.Remaining)
```

### 多维度限流检查

```go
ctx := context.Background()
userID := "user123"
userIP := "192.168.1.1"
endpoint := "/api/data"

allowed, reason := limiter.CheckMultiDimensional(ctx, userID, userIP, endpoint)
if !allowed {
    log.Printf("Request denied by %s", reason)
    return
}

// 请求通过所有维度的限流检查
```

### 带指标的限流检查

```go
res, err := limiter.CheckWithMetrics(ctx, key, limit)
if err != nil || res.Allowed == 0 {
    // 请求被限流，指标已自动记录
    return
}
```

### 导出 Prometheus 指标

```go
import "net/http"

// 注册指标端点
http.HandleFunc("/metrics", ratelimit.MetricsHandler)
http.ListenAndServe(":8080", nil)
```

## 降级策略

当 Redis 不可用时，限流器会根据配置自动降级：

### 降级放行（默认）

```go
limiter := ratelimit.NewRateLimiter(
    rdb,
    ratelimit.WithFallbackAllow(true),
)
// Redis 故障时，所有请求都会放行
```

### 降级拒绝

```go
limiter := ratelimit.NewRateLimiter(
    rdb,
    ratelimit.WithFallbackAllow(false),
    ratelimit.WithFallbackRetryAfter(5*time.Second),
)
// Redis 故障时，所有请求都会被拒绝
```

## 指标说明

导出的 Prometheus 指标：

- `ratelimit_total_requests`: 总请求数
- `ratelimit_allowed_requests`: 通过的请求数
- `ratelimit_rejected_requests`: 被拒绝的请求数
- `ratelimit_redis_errors`: Redis 错误次数
- `ratelimit_fallback_activated`: 降级激活次数

## 最佳实践

1. **合理设置超时时间**：根据网络延迟和业务需求调整 `WithRedisTimeout`
2. **选择合适的降级策略**：关键业务建议降级拒绝，非关键业务可降级放行
3. **监控指标**：持续监控 `ratelimit_redis_errors` 和 `ratelimit_fallback_activated`
4. **限流维度**：根据业务场景选择合适的限流维度组合
5. **限流阈值**：根据系统容量和业务需求合理设置各维度的限流阈值
