package models

import (
	"time"

	"gorm.io/gorm"
)

type UserBasic struct {
	gorm.Model
	Name           string     `gorm:"column:name" json:"name"`
	PassWord       string     `gorm:"column:password" json:"-"`
	Phone          string     `gorm:"column:phone" json:"phone" valid:"matches(^1[3-9]{1}\\d{9}$)"`
	Avatar         string     `gorm:"column:avatar" json:"avatar"` //头像
	Sex            int32      `gorm:"column:sex;default:3;comment:性别 1男 2女 3未知" json:"sex"`
	Status         int32      `gorm:"column:status;default:1;comment:账号状态 1正常 2禁用" json:"status"`
	ClientIp       string     `gorm:"column:client_ip" json:"client_ip"`
	LastActiveTime *time.Time `gorm:"column:last_active_time" json:"last_active_time"` // 最后活跃时间
	DeviceInfo     string     `gorm:"column:device_info" json:"device_info"`
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}
