package service

import (
	"errors"
	"fmt"
	"time"

	"ginchat/cache"
	"ginchat/config"
	"ginchat/models"

	"gorm.io/gorm"
)

// MessageService 消息业务逻辑
type MessageService struct{}

// SendMsg 发送单聊消息：校验好友 → 找到/创建房间 → 落库
func (s *MessageService) SendMsg(fromUserID, toUserID int64, content string, msgType int) (*models.Message, error) {
	// ① 校验双向好友关系（先查 Redis，miss 回查 MySQL）
	if !s.isMutualFriend(fromUserID, toUserID) {
		return nil, errors.New("对方不是你的好友")
	}

	// ② 找到/创建单聊房间
	roomID, err := s.GetOrCreateFriendRoom(fromUserID, toUserID)
	if err != nil {
		return nil, err
	}

	// ③ 入管道（异步：推送 + 攒批落库）
	msg := models.Message{
		RoomID:     roomID,
		FromUserID: fromUserID,
		Content:    content,
		Type:       msgType,
	}
	msg.CreatedAt = time.Now() // 手动带时间：异步推送时也要有发送时间
	msgCh <- messageTask{Msg: msg, ToUserIDs: []int64{toUserID}}
	return &msg, nil
}

// isMutualFriend 双向好友校验（Redis 缓存 + MySQL 兜底）
func (s *MessageService) isMutualFriend(fromUID, toUID int64) bool {
	// 先查 Redis 缓存
	if cache.IsFriend(fromUID, toUID) && cache.IsFriend(toUID, fromUID) {
		return true
	}
	// Redis miss（可能重启丢了），回查 MySQL
	var count int64
	config.Global.DB.Model(&models.UserFriend{}).
		Where("(uid = ? AND friend_uid = ? AND is_deleted = ?) OR (uid = ? AND friend_uid = ? AND is_deleted = ?)",
			fromUID, toUID, false, toUID, fromUID, false).
		Count(&count)
	if count == 2 {
		cache.AddFriend(fromUID, toUID) // 回填缓存
		return true
	}
	return false
}

// GetOrCreateFriendRoom 找到或创建单聊房间（room_key 唯一，保证同一对好友只有一个房间）
func (s *MessageService) GetOrCreateFriendRoom(uid1, uid2 int64) (int64, error) {
	// 排序，小的在前，保证 room_key 一致
	a, b := uid1, uid2
	if a > b {
		a, b = b, a
	}
	roomKey := fmt.Sprintf("%d_%d", a, b)

	var rf models.RoomFriend
	err := config.Global.DB.Where("room_key = ?", roomKey).First(&rf).Error
	if err == nil {
		return rf.RoomID, nil // 房间已存在
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err // 查询出错
	}

	// 房间不存在，创建 room + room_friend
	room := models.Room{Type: 2} // 2 = 单聊
	if err := config.Global.DB.Create(&room).Error; err != nil {
		return 0, err
	}
	rf = models.RoomFriend{
		RoomID:  int64(room.ID),
		UID1:    a,
		UID2:    b,
		RoomKey: roomKey,
	}
	if err := config.Global.DB.Create(&rf).Error; err != nil {
		return 0, err
	}
	return int64(room.ID), nil
}

// GetHistory 拉取会话的历史消息（游标分页，倒序）
func (s *MessageService) GetHistory(uid, roomID, cursor int64, limit int) ([]models.Message, bool, error) {
	if limit <= 0 {
		limit = 20
	}

	// 权限校验：单聊查 room_friend，群聊查 group_member
	var room models.Room
	if err := config.Global.DB.First(&room, roomID).Error; err != nil {
		return nil, false, errors.New("会话不存在")
	}
	if room.Type == 1 { // 群聊
		var member models.GroupMember
		if err := config.Global.DB.Where("group_id = ? AND uid = ?", roomID, uid).First(&member).Error; err != nil {
			return nil, false, errors.New("无权访问该会话")
		}
	} else { // 单聊
		var rf models.RoomFriend
		if err := config.Global.DB.Where("room_id = ? AND (uid1 = ? OR uid2 = ?)", roomID, uid, uid).First(&rf).Error; err != nil {
			return nil, false, errors.New("无权访问该会话")
		}
	}

	query := config.Global.DB.Where("room_id = ?", roomID)
	if cursor > 0 {
		query = query.Where("id < ?", cursor) // 游标：查更早的消息
	}

	var msgs []models.Message
	if err := query.Order("id DESC").Limit(limit + 1).Find(&msgs).Error; err != nil {
		return nil, false, err
	}

	// 多查一条，判断是否还有更早的消息
	hasMore := len(msgs) > limit
	if hasMore {
		msgs = msgs[:limit]
	}
	return msgs, hasMore, nil
}
