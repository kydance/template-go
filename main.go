package main

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/kydenul/log"
	"github.com/kydenul/template-go/internal/middleware"
	"github.com/kydenul/template-go/internal/stores"
	ratelimit "github.com/kydenul/template-go/pkg/rate_limit"
)

func init() {
	opt, err := log.LoadFromFile(filepath.Join(".", "config", "config.yaml"))
	if err != nil {
		panic(fmt.Sprintf("Failed to load log config from file: %v", err))
	}
	log.NewLog(opt)

	err = stores.RedisInit(filepath.Join(".", "config"))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	limiter := ratelimit.NewRateLimiter(stores.Rdb,
		ratelimit.WithGlobalQPS(300),
		ratelimit.WithIPRatePerMinute(60*200),
		ratelimit.WithUserRatePerMinute(60*200),
	)

	r := gin.New()
	r.Use(middleware.Logger(), middleware.Recovery())

	r.GET("/health", func(c *gin.Context) {
		if err := stores.Rdb.Ping(context.Background()).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "Redis unhealthy",
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "OK",
		})
	})

	r.GET("/metrics", func(c *gin.Context) { ratelimit.MetricsHandler(c.Writer, c.Request) })

	r.GET("/api/data", func(c *gin.Context) {
		userID := c.Query("X-User-ID")
		userIP := c.RemoteIP()

		allowed, reason := limiter.CheckMultiDimensional(c, userID, userIP, "api/data")
		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message": "Rate limit exceeded: " + reason,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Success",
		})
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

// ab -n 1000 -c 100 'http://127.0.0.1:8080/api/data'
