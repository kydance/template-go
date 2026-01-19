# Rate Limiter

基于 Redis 的高性能分布式限流器，使用 GCRA (Generic Cell Rate Algorithm) 算法实现，支持多维度限流和降级策略。

## 特性

- **GCRA 算法**：使用通用信元速率算法，提供精确的限流控制
- **多维度限流**：支持全局、IP、用户三个维度的限流
- **降级策略**：Redis 故障时支持配置降级行为（放行/拒绝）
- **健康检查**：内置 Redis 健康检查机制
- **指标监控**：内置 Prometheus 指标导出，支持维度标签
- **可选开关**：所有功能模块均可独立开关，按需启用
- **Option 模式**：灵活的配置方式，支持按需自定义
- **Lua 脚本原子操作**：限流逻辑在 Redis 端原子执行，保证分布式一致性
- **Redis 7.0+ 兼容**：移除了 `redis.replicate_commands()` 以兼容新版 Redis

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
    ratelimit.WithEnableMetrics(true), // 启用 Metrics 监控
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
| `WithEnableGlobalRateLimit(bool)` | 是否启用全局限流 | true |
| `WithEnableIPRateLimit(bool)` | 是否启用 IP 限流 | true |
| `WithEnableUserRateLimit(bool)` | 是否启用用户限流 | true |
| `WithEnableHealthCheck(bool)` | 是否启用 Redis 健康检查 | true |
| `WithEnableMetrics(bool)` | 是否启用 Metrics 监控 | false |

## 核心类型

### Limit

限流参数定义：

```go
type Limit struct {
    Rate   int           // 单位时间内允许的事件数
    Burst  int           // 最大突发容量（令牌桶容量）
    Period time.Duration // 时间周期
}
```

提供便捷的构造函数：

```go
ratelimit.PerSecond(rate int) Limit  // 每秒 rate 次，burst = rate
ratelimit.PerMinute(rate int) Limit  // 每分钟 rate 次，burst = rate
ratelimit.PerHour(rate int) Limit    // 每小时 rate 次，burst = rate
```

### Result

限流检查结果：

```go
type Result struct {
    Allowed    int           // 允许的事件数（n 或 0）
    Remaining  int           // 当前窗口内剩余配额
    RetryAfter time.Duration // 重试等待时间（-1 表示无需重试）
    ResetAfter time.Duration // 限流器重置时间
}
```

## 使用示例

### 基本限流

```go
ctx := context.Background()
key := "api:user:123"
limit := ratelimit.PerSecond(10) // 每秒 10 次

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

### 底层 Limiter 直接使用

如果只需要简单的限流功能，可以直接使用底层的 `Limiter`：

```go
// 创建底层限流器
l := ratelimit.NewLimiter(rdb)

// 检查是否允许 1 个事件
res, err := l.Allow(ctx, "mykey", ratelimit.PerSecond(100))

// 检查是否允许 n 个事件
res, err := l.AllowN(ctx, "mykey", ratelimit.PerSecond(100), 5)

// 检查最多允许 n 个事件（可能返回小于 n 的值）
res, err := l.AllowAtMost(ctx, "mykey", ratelimit.PerSecond(100), 10)

// 重置限流器
err := l.Reset(ctx, "mykey")
```

### 启用 Metrics 监控

```go
// 创建限流器时启用 metrics
limiter := ratelimit.NewRateLimiter(
    rdb,
    ratelimit.WithEnableMetrics(true), // 启用监控
)

// CheckMultiDimensional 会自动收集 metrics
allowed, reason := limiter.CheckMultiDimensional(ctx, userID, userIP, endpoint)
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

**按维度统计的指标（带 dimension 标签）：**
- `ratelimit_total_requests{dimension="global|ip|user"}`: 各维度总请求数
- `ratelimit_allowed_requests{dimension="global|ip|user"}`: 各维度通过的请求数
- `ratelimit_rejected_requests{dimension="global|ip|user"}`: 各维度被拒绝的请求数

**全局统计指标：**
- `ratelimit_redis_errors`: Redis 错误次数
- `ratelimit_fallback_activated`: 降级激活次数

### Metrics 示例输出

```
# HELP ratelimit_total_requests Total rate limit checks by dimension
# TYPE ratelimit_total_requests counter
ratelimit_total_requests{dimension="global"} 1234
ratelimit_total_requests{dimension="ip"} 1234
ratelimit_total_requests{dimension="user"} 890

# HELP ratelimit_allowed_requests Allowed requests by dimension
# TYPE ratelimit_allowed_requests counter
ratelimit_allowed_requests{dimension="global"} 1200
ratelimit_allowed_requests{dimension="ip"} 1150
ratelimit_allowed_requests{dimension="user"} 850

# HELP ratelimit_rejected_requests Rejected requests by dimension
# TYPE ratelimit_rejected_requests counter
ratelimit_rejected_requests{dimension="global"} 34
ratelimit_rejected_requests{dimension="ip"} 84
ratelimit_rejected_requests{dimension="user"} 40

# HELP ratelimit_redis_errors Redis errors
# TYPE ratelimit_redis_errors counter
ratelimit_redis_errors 5

# HELP ratelimit_fallback_activated Fallback activations
# TYPE ratelimit_fallback_activated counter
ratelimit_fallback_activated 5
```

## 算法说明

本限流器使用 GCRA (Generic Cell Rate Algorithm) 算法，也称为"漏桶"算法的变种。其核心特点：

1. **时间一致性**：使用 Redis 服务器时间 (`TIME` 命令) 而非客户端时间，保证分布式环境下的时钟一致性
2. **原子操作**：所有限流逻辑通过 Lua 脚本在 Redis 端原子执行，避免竞态条件
3. **精确控制**：基于"理论到达时间"(TAT) 计算，提供平滑的流量控制
4. **自动过期**：Redis key 自动设置过期时间，无需手动清理

### Key 命名规则

限流器使用的 Redis key 前缀为 `rate:`，完整 key 格式：

- 全局限流：`rate:global:ratelimit:{endpoint}`
- IP 限流：`rate:ratelimit:ip:{ip}`
- 用户限流：`rate:ratelimit:user:{userID}:{endpoint}`

## 最佳实践

1. **合理设置超时时间**：根据网络延迟和业务需求调整 `WithRedisTimeout`
2. **选择合适的降级策略**：关键业务建议降级拒绝，非关键业务可降级放行
3. **监控指标**：持续监控 `ratelimit_redis_errors` 和 `ratelimit_fallback_activated`
4. **限流维度**：根据业务场景选择合适的限流维度组合
5. **限流阈值**：根据系统容量和业务需求合理设置各维度的限流阈值
6. **按需启用 Metrics**：仅在需要监控时启用 `WithEnableMetrics(true)`，避免不必要的性能开销
7. **使用 Prometheus 告警**：为关键指标设置告警规则，及时发现异常

## 依赖

- [github.com/redis/go-redis/v9](https://github.com/redis/go-redis) - Redis 客户端
