package cache

import (
	"context"
	"fmt"
	"log"

	"ginchat/config"
)

// failMsgKey 落库失败消息的死信队列 key
const failMsgKey = "fail_msg"

func offlineKey(userID int64) string {
	return fmt.Sprintf("offline:%d", userID)
}

// SaveOfflineMsg 存离线消息（失败降级）
func SaveOfflineMsg(userID int64, msgJSON []byte) {
	if err := config.Global.Red.RPush(context.Background(), offlineKey(userID), msgJSON).Err(); err != nil {
		log.Printf("[cache] SaveOfflineMsg 失败: userID=%d, err=%v", userID, err)
	}
}

// GetOfflineMsg 拉取并清空离线消息
func GetOfflineMsg(userID int64) []string {
	ctx := context.Background()
	list, err := config.Global.Red.LRange(ctx, offlineKey(userID), 0, -1).Result()
	if err != nil {
		log.Printf("[cache] GetOfflineMsg 查询失败: userID=%d, err=%v", userID, err)
		return nil
	}
	config.Global.Red.Del(ctx, offlineKey(userID))
	return list
}

// SaveFailMsg 落库失败的消息（死信队列，定时重试）
func SaveFailMsg(msgJSON []byte) {
	if err := config.Global.Red.RPush(context.Background(), failMsgKey, msgJSON).Err(); err != nil {
		log.Printf("[cache] SaveFailMsg 失败: err=%v", err)
	}
}

// GetFailMsg 获取所有失败消息
func GetFailMsg() []string {
	list, err := config.Global.Red.LRange(context.Background(), failMsgKey, 0, -1).Result()
	if err != nil {
		log.Printf("[cache] GetFailMsg 查询失败: err=%v", err)
		return nil
	}
	return list
}

// RemoveFailMsg 删除一条失败消息（重试成功后）
func RemoveFailMsg(msgJSON string) {
	if err := config.Global.Red.LRem(context.Background(), failMsgKey, 0, msgJSON).Err(); err != nil {
		log.Printf("[cache] RemoveFailMsg 失败: err=%v", err)
	}
}
