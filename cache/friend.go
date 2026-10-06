package cache

import (
	"context"
	"fmt"

	"ginchat/config"
)

// friendKey 好友关系缓存 key
func friendKey(uid int64) string {
	return fmt.Sprintf("friend:%d", uid)
}

// AddFriend 加好友：双向写入缓存
func AddFriend(uid1, uid2 int64) {
	ctx := context.Background()
	config.Global.Red.SAdd(ctx, friendKey(uid1), uid2)
	config.Global.Red.SAdd(ctx, friendKey(uid2), uid1)
}

// RemoveFriend 删好友：双向移除缓存
func RemoveFriend(uid1, uid2 int64) {
	ctx := context.Background()
	config.Global.Red.SRem(ctx, friendKey(uid1), uid2)
	config.Global.Red.SRem(ctx, friendKey(uid2), uid1)
}

// IsFriend 判断 uid1 的好友里是否有 uid2（查 Redis）
func IsFriend(uid1, uid2 int64) bool {
	return config.Global.Red.SIsMember(context.Background(), friendKey(uid1), uid2).Val()
}
