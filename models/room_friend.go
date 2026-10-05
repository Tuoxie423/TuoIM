package models

import "gorm.io/gorm"

// RoomFriend 单聊会话表
type RoomFriend struct {
	gorm.Model
	RoomID  int64  `gorm:"column:room_id;not null;comment:房间id" json:"room_id"`
	UID1    int64  `gorm:"column:uid1;not null;comment:较小的uid" json:"uid1"`
	UID2    int64  `gorm:"column:uid2;not null;comment:较大的uid" json:"uid2"`
	RoomKey string `gorm:"column:room_key;not null;uniqueIndex;comment:uid1_uid2 排序拼接，防重复建房间" json:"room_key"`
	Status  int    `gorm:"column:status;default:0;comment:房间状态 0正常 1禁用" json:"status"`
}

func (RoomFriend) TableName() string {
	return "room_friend"
}
