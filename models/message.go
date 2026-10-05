package models

import "gorm.io/gorm"

type Message struct {
	gorm.Model
	RoomID     int64  `gorm:"column:room_id;not null;comment:会话id" json:"room_id"`
	FromUserID int64  `gorm:"column:from_user_id;not null;comment:发送者id" json:"from_user_id"`
	Content    string `gorm:"column:content" json:"content"`
	Type       int    `gorm:"column:type;default:1;comment:消息类型 1文本 2图片" json:"type"`
}

func (Message) TableName() string {
	return "message"
}
