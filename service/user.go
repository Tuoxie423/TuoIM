package service

import (
	"errors"

	"ginchat/config"
	"ginchat/models"
	"ginchat/utils"
)

// UserService 用户业务逻辑
type UserService struct{}

// Register 用户注册
func (s *UserService) Register(name, password, phone string) (*models.UserBasic, error) {
	if name == "" || password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}
	if len(password) < 6 {
		return nil, errors.New("密码长度不能少于6位")
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
// 用户不存在和密码错误统一返回，避免手机号枚举
func (s *UserService) Login(phone, password string) (*models.UserBasic, error) {
	var user models.UserBasic
	if err := config.Global.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if !utils.CheckPasswordHash(password, user.PassWord) {
		return nil, errors.New("用户名或密码错误")
	}
	return &user, nil
}

// GetUserByID 按 ID 查询用户
func (s *UserService) GetUserByID(id int64) (*models.UserBasic, error) {
	var user models.UserBasic
	if err := config.Global.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户资料（昵称/头像/性别）
func (s *UserService) UpdateUser(id int64, name, avatar string, sex int32) error {
	updates := map[string]any{}
	if name != "" {
		updates["name"] = name
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if sex != 0 {
		updates["sex"] = sex
	}
	if len(updates) == 0 {
		return errors.New("没有要更新的字段")
	}
	return config.Global.DB.Model(&models.UserBasic{}).Where("id = ?", id).Updates(updates).Error
}
