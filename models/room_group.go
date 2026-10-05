package models

import "gorm.io/gorm"

// RoomGroup 群聊会话表
type RoomGroup struct {
	gorm.Model
	RoomID       int64  `gorm:"column:room_id;not null;comment:房间id" json:"room_id"`
	Name         string `gorm:"column:name;not null;comment:群名称" json:"name"`
	Avatar       string `gorm:"column:avatar;comment:群头像" json:"avatar"`
	DeleteStatus int    `gorm:"column:delete_status;default:0;comment:逻辑删除 0正常 1删除" json:"delete_status"`
}

func (RoomGroup) TableName() string {
	return "room_group"
}
