package main

import (
	"fmt"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/kydenul/log"
	"github.com/kydenul/template-go/internal/middleware"
	"github.com/kydenul/template-go/internal/service"
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

	baseSvr, err := service.NewBaseServer(r, stores.Rdb, limiter)
	if err != nil {
		log.Fatal(err)
	}

	svr := service.NewServer(*baseSvr)
	svr.InitRouter()

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

// ab -n 1000 -c 100 'http://127.0.0.1:8080/api/data'
