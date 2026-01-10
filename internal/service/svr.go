package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kydenul/log"
	"github.com/kydenul/template-go/internal/pb"
	"github.com/kydenul/template-go/internal/stores"
	ratelimit "github.com/kydenul/template-go/pkg/rate_limit"
)

type Server struct {
	BaseServer
}

func NewServer(baseSvr BaseServer) *Server {
	return &Server{baseSvr}
}

func (*Server) RedisHealthHandler(c *gin.Context) {
	if err := stores.Rdb.Ping(context.Background()).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "Redis unhealthy",
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func (*Server) RateLimiterMetricsHandler(c *gin.Context) {
	ratelimit.MetricsHandler(c.Writer, c.Request)
}

func (svr *Server) APIDataHandler(c *gin.Context) {
	userID := c.Query("X-User-ID")
	userIP := c.RemoteIP()

	allowed, reason := svr.RateLimiter.CheckMultiDimensional(c, userID, userIP, "api/data")
	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"message": "Rate limit exceeded: " + reason,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func (svr *Server) UserHandler(c *gin.Context) {
	var req pb.UserRequest

	// 1. bind request
	// Content-Type: application/x-protobuf
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, svr.Err(c, http.StatusBadRequest, err.Error()))
		return
	}

	log.Infof("req: %s", req.String())

	resp := &pb.UserResponse{
		Id:    1001,
		Name:  req.Name,
		Email: "kyden@example.com",
	}

	// 2. return Protobuf
	switch c.GetHeader("Accept") {
	case "application/x-protobuf":
		log.Debugf("resp: %+v", resp)

		// Content-Type: application/x-protobuf
		c.ProtoBuf(http.StatusOK, svr.Succ(c, resp))

	case "application/json":
		log.Debugf("resp: %+v", resp)

		c.JSON(http.StatusOK, svr.Succ(c, resp))

	default:
		log.Errorf("unknown Accept header: %s", c.GetHeader("Accept"))
		c.JSON(http.StatusBadRequest, svr.Err(c, http.StatusBadRequest, "unknown Accept header"))
	}
}
