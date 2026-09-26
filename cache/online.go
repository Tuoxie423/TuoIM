package cache

import (
	"context"
	"fmt"
	"time"

	"ginchat/config"
)

const onlineKeyPrefix = "online:"

func onlineKey(userID uint) string {
	return fmt.Sprintf("%s%d", onlineKeyPrefix, userID)
}

// SetOnline 标记用户在线，带过期时间（RedisOnlineTime 单位：小时）
func SetOnline(userID uint) error {
	ttl := time.Duration(config.Global.Timeout.RedisOnlineTime) * time.Hour
	return config.Global.Red.Set(context.Background(), onlineKey(userID), 1, ttl).Err()
}

// DelOnline 移除用户在线状态
func DelOnline(userID uint) error {
	return config.Global.Red.Del(context.Background(), onlineKey(userID)).Err()
}

// IsOnline 判断用户是否在线
func IsOnline(userID uint) bool {
	exists, _ := config.Global.Red.Exists(context.Background(), onlineKey(userID)).Result()
	return exists > 0
}
