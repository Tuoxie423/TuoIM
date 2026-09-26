package initial

import (
	"ginchat/config"

	"github.com/go-redis/redis/v8"
)

func InitRedis() {
	config.Global.Red = redis.NewClient(&redis.Options{
		Addr:         config.Global.Redis.Addr,
		Password:     config.Global.Redis.Password,
		DB:           config.Global.Redis.DB,
		PoolSize:     config.Global.Redis.PoolSize,
		MinIdleConns: config.Global.Redis.MinIdleConn,
	})
}
