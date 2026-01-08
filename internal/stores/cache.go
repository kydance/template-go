package stores

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kydenul/log"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

var (
	Rdb           redis.UniversalClient
	redisInitOnce sync.Once
)

func RedisInit(configPath string) error {
	var err error
	redisInitOnce.Do(func() {
		viper.AddConfigPath(configPath)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("failed to read config: %v", err)
		}

		host := viper.GetString("Redis.host")
		port := viper.GetString("Redis.port")
		poolSize := viper.GetInt("Redis.pool_size")
		maxIdle := viper.GetInt("Redis.max_idle")

		log.Debugf("RedisClient, host: %s, port: %s, poolSize: %d, maxIdle: %d",
			host, port, poolSize, maxIdle)

		Rdb = redis.NewClient(&redis.Options{
			Addr:     host + ":" + port,
			Password: viper.GetString("Redis.password"),

			PoolSize:        poolSize,
			MaxIdleConns:    maxIdle,
			MinIdleConns:    5,                // 保持最小空闲连接，避免冷启动
			ConnMaxIdleTime: 10 * time.Minute, // 空闲连接 10 分钟后关闭
			ConnMaxLifetime: 30 * time.Minute, // 强制连接运行一段时间后必须重建，防止由于拓扑变化导致的连接黑洞

			ReadBufferSize:  1024 * 1024, // 1MiB read buffer
			WriteBufferSize: 1024 * 1024, // 1MiB write buffer

			PoolFIFO: true, // 使用FIFO模式，更快清理空闲连接
		})

		// Test Conn with retry
		var pingErr error
		maxRetries := 3
		for i := range maxRetries {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			pingErr = Rdb.Ping(ctx).Err()
			cancel()

			if pingErr == nil {
				break
			}

			log.Errorf("go-redis/v9 redis ping failed (attempt %d/%d): %s",
				i+1, maxRetries, pingErr.Error())

			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * time.Second)
			}
		}

		if pingErr != nil {
			err = fmt.Errorf("redis ping failed after %d retries: %w", maxRetries, pingErr)
		}

		log.Info("go-redis/v9 redis init ok ^_^")

		// NOTE: Pool Stats
		ticker := time.NewTicker(1 * time.Minute)
		go func() {
			for range ticker.C {
				stats := Rdb.PoolStats()
				log.Infof(
					"Redis Pool: Hits=%d Misses=%d Timeouts=%d TotalConns=%d IdleConns=%d StaleConns=%d",
					stats.Hits,
					stats.Misses,
					stats.Timeouts,
					stats.TotalConns,
					stats.IdleConns,
					stats.StaleConns,
				)
			}
		}()
	})

	return err
}
