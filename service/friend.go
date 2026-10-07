package service

import (
	"errors"

	"ginchat/cache"
	"ginchat/config"
	"ginchat/models"
)

// FriendService 好友业务逻辑
type FriendService struct{}

// ApplyInfo 好友申请信息（含申请人资料）
type ApplyInfo struct {
	ApplyID int64  `json:"apply_id"`
	UID     int64  `json:"uid"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Msg     string `json:"msg"`
	Status  int    `json:"status"`
}

// FriendInfo 好友信息
type FriendInfo struct {
	UID    int64  `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Online bool   `json:"online"` // 是否在线
}

// SearchUser 按手机号搜索用户
func (s *FriendService) SearchUser(phone string) (*models.UserBasic, error) {
	var user models.UserBasic
	if err := config.Global.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

// ApplyFriend 发起好友申请
func (s *FriendService) ApplyFriend(fromUID, toUID int64, msg string) error {
	if fromUID == toUID {
		return errors.New("不能添加自己为好友")
	}

	// 目标用户是否存在
	var target models.UserBasic
	if err := config.Global.DB.First(&target, toUID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 是否已经是好友
	var count int64
	config.Global.DB.Model(&models.UserFriend{}).
		Where("uid = ? AND friend_uid = ? AND is_deleted = ?", fromUID, toUID, false).
		Count(&count)
	if count > 0 {
		return errors.New("你们已经是好友了")
	}

	// 是否已发过待审批的申请
	var apply models.UserApply
	err := config.Global.DB.Where("uid = ? AND target_id = ? AND status = ?", fromUID, toUID, 1).
		First(&apply).Error
	if err == nil {
		return errors.New("已发送过申请，请等待对方同意")
	}

	// 写申请
	apply = models.UserApply{
		UID:      fromUID,
		TargetID: toUID,
		Msg:      msg,
		Status:   1, // 待审批
	}
	return config.Global.DB.Create(&apply).Error
}

// AgreeFriend 同意好友申请
func (s *FriendService) AgreeFriend(currentUserID, applyUserID int64) error {
	// 查待审批的申请（申请人 applyUserID → 接收人 currentUserID）
	var apply models.UserApply
	err := config.Global.DB.Where("uid = ? AND target_id = ? AND status = ?", applyUserID, currentUserID, 1).
		First(&apply).Error
	if err != nil {
		return errors.New("好友申请不存在或已处理")
	}

	tx := config.Global.DB.Begin()

	// 更新申请状态 → 已同意
	if err := tx.Model(&models.UserApply{}).Where("id = ?", apply.ID).
		Update("status", 2).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 双向建好友关系（FirstOrCreate + Assign 处理"删过好友重新加"的复活）
	pairs := [][2]int64{{currentUserID, applyUserID}, {applyUserID, currentUserID}}
	for _, p := range pairs {
		if err := tx.Where("uid = ? AND friend_uid = ?", p[0], p[1]).
			Assign(models.UserFriend{IsDeleted: false}).
			FirstOrCreate(&models.UserFriend{UID: p[0], FriendUID: p[1]}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	cache.AddFriend(currentUserID, applyUserID) // 写好友缓存
	return nil
}

// RejectFriend 拒绝好友申请
func (s *FriendService) RejectFriend(currentUserID, applyUserID int64) error {
	result := config.Global.DB.Model(&models.UserApply{}).
		Where("uid = ? AND target_id = ? AND status = ?", applyUserID, currentUserID, 1).
		Update("status", 3)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("好友申请不存在或已处理")
	}
	return nil
}

// DeleteFriend 删除好友（单向删除：只删自己这边）
func (s *FriendService) DeleteFriend(uid, friendUID int64) error {
	result := config.Global.DB.Model(&models.UserFriend{}).
		Where("uid = ? AND friend_uid = ?", uid, friendUID).
		Update("is_deleted", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("你们还不是好友")
	}
	cache.RemoveFriend(uid, friendUID) // 清缓存
	return nil
}

// GetApplyList 查看好友申请列表（别人发给我的），并标记为已读
func (s *FriendService) GetApplyList(uid int64) ([]ApplyInfo, error) {
	var applies []models.UserApply
	if err := config.Global.DB.Where("target_id = ?", uid).
		Order("created_at DESC").Find(&applies).Error; err != nil {
		return nil, err
	}
	if len(applies) == 0 {
		return []ApplyInfo{}, nil
	}

	// 收集申请人 uid，批量查用户信息
	uids := make([]int64, 0, len(applies))
	for _, a := range applies {
		uids = append(uids, a.UID)
	}
	var users []models.UserBasic
	config.Global.DB.Where("id IN ?", uids).Find(&users)
	userMap := make(map[int64]models.UserBasic, len(users))
	for _, u := range users {
		userMap[int64(u.ID)] = u
	}

	// 组合响应
	result := make([]ApplyInfo, 0, len(applies))
	for _, a := range applies {
		info := ApplyInfo{
			ApplyID: int64(a.ID),
			UID:     a.UID,
			Msg:     a.Msg,
			Status:  a.Status,
		}
		if u, ok := userMap[a.UID]; ok {
			info.Name = u.Name
			info.Avatar = u.Avatar
		}
		result = append(result, info)
	}

	// 标记为已读
	config.Global.DB.Model(&models.UserApply{}).
		Where("target_id = ? AND read_status = ?", uid, 1).
		Update("read_status", 2)

	return result, nil
}

// GetFriendList 好友列表
func (s *FriendService) GetFriendList(uid int64) ([]FriendInfo, error) {
	var friends []models.UserFriend
	if err := config.Global.DB.Where("uid = ? AND is_deleted = ?", uid, false).
		Find(&friends).Error; err != nil {
		return nil, err
	}
	if len(friends) == 0 {
		return []FriendInfo{}, nil
	}

	uids := make([]int64, 0, len(friends))
	for _, f := range friends {
		uids = append(uids, f.FriendUID)
	}
	var users []models.UserBasic
	config.Global.DB.Where("id IN ?", uids).Find(&users)
	userMap := make(map[int64]models.UserBasic, len(users))
	for _, u := range users {
		userMap[int64(u.ID)] = u
	}

	result := make([]FriendInfo, 0, len(friends))
	for _, f := range friends {
		info := FriendInfo{UID: f.FriendUID, Online: cache.IsOnline(f.FriendUID)}
		if u, ok := userMap[f.FriendUID]; ok {
			info.Name = u.Name
			info.Avatar = u.Avatar
		}
		result = append(result, info)
	}
	return result, nil
}

// UnreadApplyNum 未读申请数
func (s *FriendService) UnreadApplyNum(uid int64) (int64, error) {
	var count int64
	if err := config.Global.DB.Model(&models.UserApply{}).
		Where("target_id = ? AND read_status = ?", uid, 1).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
