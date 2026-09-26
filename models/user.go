package models

import (
	"time"

	"gorm.io/gorm"
)

type UserBasic struct {
	gorm.Model
	Name          string     `gorm:"column:name" json:"name"`
	PassWord      string     `gorm:"column:password" json:"password"`
	Phone         string     `gorm:"column:phone" json:"phone" valid:"matches(^1[3-9]{1}\\d{9}$)"`
	Email         string     `gorm:"column:email" json:"email" valid:"email"`
	Avatar        string     `gorm:"column:avatar" json:"avatar"` //头像
	Identity      string     `gorm:"column:identity" json:"identity"`
	ClientIp      string     `gorm:"column:client_ip" json:"client_ip"`
	ClientPort    string     `gorm:"column:client_port" json:"client_port"`
	Salt          string     `gorm:"column:salt" json:"salt"` //盐
	LoginTime     *time.Time `gorm:"column:login_time" json:"login_time"`
	HeartbeatTime *time.Time `gorm:"column:heartbeat_time" json:"heartbeat_time"`
	LoginOutTime  *time.Time `gorm:"column:login_out_time" json:"login_out_time"`
	IsLogout      bool       `gorm:"column:is_logout" json:"is_logout"`
	DeviceInfo    string     `gorm:"column:device_info" json:"device_info"`
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}
