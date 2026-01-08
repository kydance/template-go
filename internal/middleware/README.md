# Middleware Package

这个包提供了与 `kydenul/log` 集成的 Gin 中间件。

## Logger Middleware

自定义日志中间件，记录每个 HTTP 请求的详细信息。

### 特性

- **自动日志级别**: 根据 HTTP 状态码自动选择日志级别
  - `2xx-3xx`: Info 级别
  - `4xx`: Warn 级别
  - `5xx`: Error 级别
- **详细信息**: 记录状态码、延迟、客户端IP、方法、响应大小和路径
- **可配置跳过路径**: 支持跳过指定路径（如健康检查端点）
- **错误追踪**: 自动记录 Gin 错误信息

### 使用方式

#### 基本用法

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/kydenul/template-go/internal/middleware"
)

func main() {
    r := gin.New()

    // 使用自定义日志中间件
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())

    // ... 注册路由
}
```

#### 高级用法（跳过特定路径）

```go
r := gin.New()

// 配置跳过健康检查端点的日志
r.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    SkipPaths: []string{"/health", "/metrics"},
}))
r.Use(middleware.Recovery())
```

### 日志格式

```
[GIN] 200 |      1.234ms |      127.0.0.1 | GET     |      42 | /api/data
[GIN] 404 |      0.123ms |      127.0.0.1 | POST    |     128 | /api/notfound | 404 page not found
[GIN] 500 |     10.234ms |      127.0.0.1 | PUT     |      56 | /api/error | Internal Server Error
```

格式说明：
- **状态码**: HTTP 响应状态码
- **延迟**: 请求处理时间
- **客户端IP**: 请求来源 IP
- **方法**: HTTP 方法（GET/POST/PUT/DELETE等）
- **响应大小**: 响应体大小（字节）
- **路径**: 请求路径（包含查询参数）
- **错误信息**: 如果有错误，显示错误详情

## Recovery Middleware

自定义恢复中间件，捕获 panic 并记录到日志。

### 特性

- **Panic 捕获**: 捕获处理器中的 panic
- **日志记录**: 使用 `kydenul/log` 记录 panic 详情
- **优雅降级**: 返回 500 状态码而不是崩溃

### 使用方式

```go
r := gin.New()
r.Use(middleware.Recovery())
```

### 示例输出

当发生 panic 时：
```
[GIN] panic recovered: runtime error: index out of range [0] with length 0
```

## 为什么不使用 gin.Default()？

`gin.Default()` 使用 Gin 内置的 logger 和 recovery 中间件，但它们不会与 `kydenul/log` 集成。

通过使用 `gin.New()` 和自定义中间件，我们可以：
1. 统一日志输出到 `kydenul/log`
2. 保持日志格式一致
3. 利用 `kydenul/log` 的文件轮转、日志级别等特性
4. 更好的日志管理和监控

## 完整示例

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/kydenul/log"
    "github.com/kydenul/template-go/internal/middleware"
)

func main() {
    // 初始化日志
    log.NewLog(&log.Options{
        Level:  "info",
        Format: "console",
    })

    // 创建 Gin 实例
    r := gin.New()

    // 使用自定义中间件
    r.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
        SkipPaths: []string{"/health"},
    }))
    r.Use(middleware.Recovery())

    // 注册路由
    r.GET("/api/hello", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello World"})
    })

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    r.Run(":8080")
}
```
