package cache

import (
	"context"
	"fmt"
	"log"

	"ginchat/config"
)

// friendKey 好友关系缓存 key
func friendKey(uid int64) string {
	return fmt.Sprintf("friend:%d", uid)
}

// AddFriend 加好友：双向写入缓存（失败降级：记日志）
func AddFriend(uid1, uid2 int64) {
	ctx := context.Background()
	if err := config.Global.Red.SAdd(ctx, friendKey(uid1), uid2).Err(); err != nil {
		log.Printf("[cache] AddFriend 失败: uid1=%d, uid2=%d, err=%v", uid1, uid2, err)
	}
	if err := config.Global.Red.SAdd(ctx, friendKey(uid2), uid1).Err(); err != nil {
		log.Printf("[cache] AddFriend 失败: uid1=%d, uid2=%d, err=%v", uid1, uid2, err)
	}
}

// RemoveFriend 删好友：双向移除缓存（失败降级）
func RemoveFriend(uid1, uid2 int64) {
	ctx := context.Background()
	if err := config.Global.Red.SRem(ctx, friendKey(uid1), uid2).Err(); err != nil {
		log.Printf("[cache] RemoveFriend 失败: uid1=%d, uid2=%d, err=%v", uid1, uid2, err)
	}
	if err := config.Global.Red.SRem(ctx, friendKey(uid2), uid1).Err(); err != nil {
		log.Printf("[cache] RemoveFriend 失败: uid1=%d, uid2=%d, err=%v", uid1, uid2, err)
	}
}

// IsFriend 判断 uid1 的好友里是否有 uid2（查询失败降级为 false）
func IsFriend(uid1, uid2 int64) bool {
	res, err := config.Global.Red.SIsMember(context.Background(), friendKey(uid1), uid2).Result()
	if err != nil {
		log.Printf("[cache] IsFriend 查询失败: uid1=%d, uid2=%d, err=%v", uid1, uid2, err)
		return false
	}
	return res
}
