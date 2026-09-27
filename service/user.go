package service

import (
	"errors"

	"ginchat/config"
	"ginchat/models"
	"ginchat/utils"
)

// UserService 用户业务逻辑
type UserService struct{}

func (s *UserService) GetUser() ([]*models.UserBasic, error) {
	var users []*models.UserBasic
	if err := config.Global.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil

}

// Register 用户注册
func (s *UserService) Register(name, password, phone string) (*models.UserBasic, error) {
	if name == "" || password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}
	if phone == "" {
		return nil, errors.New("手机号不能为空")
	}

	// 手机号查重（生产环境建议配合数据库唯一索引做双保险）
	var count int64
	if err := config.Global.DB.Model(&models.UserBasic{}).Where("phone = ?", phone).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("该手机号已注册")
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := models.UserBasic{
		Name:     name,
		PassWord: hash,
		Phone:    phone,
	}
	if err := config.Global.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Login 用户登录（手机号 + 密码）
func (s *UserService) Login(phone, password string) (*models.UserBasic, error) {
	var user models.UserBasic
	if err := config.Global.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	if !utils.CheckPasswordHash(password, user.PassWord) {
		return nil, errors.New("密码错误")
	}
	return &user, nil
}
