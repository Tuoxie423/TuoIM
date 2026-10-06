package models

import "gorm.io/gorm"

// UserFriend 好友关系表（双向存储：A↔B 建两条记录）
type UserFriend struct {
	gorm.Model
	UID       int64 `gorm:"column:uid;not null;uniqueIndex:idx_uid_friend;comment:用户id" json:"uid"`
	FriendUID int64 `gorm:"column:friend_uid;not null;uniqueIndex:idx_uid_friend;comment:好友uid" json:"friend_uid"`
	IsDeleted bool  `gorm:"column:is_deleted;default:false;comment:是否已删除" json:"is_deleted"`
}

func (UserFriend) TableName() string {
	return "user_friend"
}
