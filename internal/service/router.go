package service

func (svr *Server) InitRouter() {
	svr.Engine.GET("/health", svr.RedisHealthHandler)
	svr.Engine.GET("/metrics", svr.RateLimiterMetricsHandler)
	svr.Engine.GET("/api/data", svr.APIDataHandler)
	svr.Engine.GET("/user", svr.UserHandler)
}
