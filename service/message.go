package service

import (
	"errors"
	"fmt"

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
	roomID, err := s.getOrCreateFriendRoom(fromUserID, toUserID)
	if err != nil {
		return nil, err
	}

	// ③ 落库
	msg := models.Message{
		RoomID:     roomID,
		FromUserID: fromUserID,
		Content:    content,
		Type:       msgType,
	}
	if err := config.Global.DB.Create(&msg).Error; err != nil {
		return nil, err
	}
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

// getOrCreateFriendRoom 找到或创建单聊房间（room_key 唯一，保证同一对好友只有一个房间）
func (s *MessageService) getOrCreateFriendRoom(uid1, uid2 int64) (int64, error) {
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
