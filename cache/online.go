package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"ginchat/config"
)

const onlineKeyPrefix = "online:"

func onlineKey(userID int64) string {
	return fmt.Sprintf("%s%d", onlineKeyPrefix, userID)
}

// SetOnline 标记用户在线，带过期时间（RedisOnlineTime 单位：小时）
// 缓存失败降级：记日志，不影响主流程
func SetOnline(userID int64) {
	ttl := time.Duration(config.Global.Timeout.RedisOnlineTime) * time.Hour
	if err := config.Global.Red.Set(context.Background(), onlineKey(userID), 1, ttl).Err(); err != nil {
		log.Printf("[cache] SetOnline 失败: userID=%d, err=%v", userID, err)
	}
}

// DelOnline 移除用户在线状态
func DelOnline(userID int64) {
	if err := config.Global.Red.Del(context.Background(), onlineKey(userID)).Err(); err != nil {
		log.Printf("[cache] DelOnline 失败: userID=%d, err=%v", userID, err)
	}
}

// IsOnline 判断用户是否在线（查询失败降级为「不在线」）
func IsOnline(userID int64) bool {
	exists, err := config.Global.Red.Exists(context.Background(), onlineKey(userID)).Result()
	if err != nil {
		log.Printf("[cache] IsOnline 查询失败: userID=%d, err=%v", userID, err)
		return false
	}
	return exists > 0
}
