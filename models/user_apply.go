package models

import "gorm.io/gorm"

// UserApply 好友申请表
type UserApply struct {
	gorm.Model
	UID        int64  `gorm:"column:uid;not null;uniqueIndex:idx_uid_target;comment:申请人uid" json:"uid"`
	TargetID   int64  `gorm:"column:target_id;not null;uniqueIndex:idx_uid_target;comment:接收人uid" json:"target_id"`
	Msg        string `gorm:"column:msg;comment:申请信息" json:"msg"`
	Status     int    `gorm:"column:status;default:1;comment:申请状态 1待审批 2同意 3拒绝" json:"status"`
	ReadStatus int    `gorm:"column:read_status;default:1;comment:阅读状态 1未读 2已读" json:"read_status"`
}

func (UserApply) TableName() string {
	return "user_apply"
}
