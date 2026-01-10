package service

import (
	"github.com/gin-gonic/gin"
	ratelimit "github.com/kydenul/template-go/pkg/rate_limit"
	"github.com/redis/go-redis/v9"
)

type BaseServer struct {
	Engine *gin.Engine

	Rdb         redis.UniversalClient
	RateLimiter *ratelimit.RateLimiter
}

func NewBaseServer(
	g *gin.Engine,
	rdb redis.UniversalClient,
	rl *ratelimit.RateLimiter,
) (*BaseServer, error) {
	return &BaseServer{
		Engine:      g,
		Rdb:         rdb,
		RateLimiter: rl,
	}, nil
}

func (*BaseServer) Err(c *gin.Context, err int, msg string) gin.H {
	return gin.H{
		"err":      err,
		"msg":      msg,
		"trace_id": c.GetHeader("X-Trace-ID"),
	}
}

func (*BaseServer) Succ(c *gin.Context, resp any) gin.H {
	return gin.H{
		"err":      "0",
		"msg":      "ok",
		"trace_id": c.GetHeader("X-Trace-ID"),
		"data":     resp,
	}
}
