package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kydenul/log"
)

// LoggerConfig 日志中间件配置
type LoggerConfig struct {
	// SkipPaths 跳过记录日志的路径（例如健康检查端点）
	SkipPaths []string
}

// Logger 返回一个自定义的日志中间件
func Logger() gin.HandlerFunc {
	return LoggerWithConfig(LoggerConfig{})
}

// LoggerWithConfig 返回一个带配置的自定义日志中间件
func LoggerWithConfig(config LoggerConfig) gin.HandlerFunc {
	skipPaths := make(map[string]bool, len(config.SkipPaths))
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 跳过指定路径
		if skipPaths[path] {
			return
		}

		// 记录请求结束时间
		latency := time.Since(start)

		// 获取客户端 IP
		clientIP := c.ClientIP()

		// 获取请求方法
		method := c.Request.Method

		// 获取状态码
		statusCode := c.Writer.Status()

		// 获取响应大小
		bodySize := c.Writer.Size()

		// 获取错误信息（如果有）
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 构建完整路径
		if raw != "" {
			path = path + "?" + raw
		}

		// 根据状态码选择日志级别
		switch {
		case statusCode >= 500:
			log.Errorf("[GIN] %3d | %13v | %15s | %-7s | %7d | %s | %s",
				statusCode,
				latency,
				clientIP,
				method,
				bodySize,
				path,
				errorMessage,
			)
		case statusCode >= 400:
			log.Warnf("[GIN] %3d | %13v | %15s | %-7s | %7d | %s | %s",
				statusCode,
				latency,
				clientIP,
				method,
				bodySize,
				path,
				errorMessage,
			)
		default:
			log.Infof("[GIN] %3d | %13v | %15s | %-7s | %7d | %s",
				statusCode,
				latency,
				clientIP,
				method,
				bodySize,
				path,
			)
		}
	}
}

// Recovery 返回一个自定义的恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("[GIN] panic recovered: %v", err)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
