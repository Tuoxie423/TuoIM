package models

import "gorm.io/gorm"

// GroupMember 群成员表
type GroupMember struct {
	gorm.Model
	GroupID int64 `gorm:"column:group_id;not null;comment:群id(room_group.id)" json:"group_id"`
	UID     int64 `gorm:"column:uid;not null;comment:成员uid" json:"uid"`
	Role    int   `gorm:"column:role;not null;comment:成员角色 1群主 2管理员 3普通成员" json:"role"`
}

func (GroupMember) TableName() string {
	return "group_member"
}
