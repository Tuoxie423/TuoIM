package models

import (
	"time"

	"gorm.io/gorm"
)

// Room 会话表（单聊、群聊统一抽象）
type Room struct {
	gorm.Model
	Type       int        `gorm:"column:type;not null;comment:房间类型 1群聊 2单聊" json:"type"`
	ActiveTime *time.Time `gorm:"column:active_time;comment:最后消息时间" json:"active_time"`
	LastMsgID  int64      `gorm:"column:last_msg_id;comment:最后一条消息id" json:"last_msg_id"`
}

func (Room) TableName() string {
	return "room"
}
