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
func (s *UserService) Register(name, password, phone, email string) (*models.UserBasic, error) {
	if name == "" || password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}

	// 简单加盐加密；生产环境建议用 bcrypt 或 scrypt
	salt := utils.MD5(phone + email)
	user := models.UserBasic{
		Name:     name,
		PassWord: utils.MD5(password + salt),
		Phone:    phone,
		Email:    email,
		Salt:     salt,
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
	if utils.MD5(password+user.Salt) != user.PassWord {
		return nil, errors.New("密码错误")
	}
	return &user, nil
}
