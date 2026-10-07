package service

import (
	"errors"
	"time"

	"ginchat/config"
	"ginchat/models"
)

// GroupService 群聊业务逻辑
type GroupService struct{}

// 成员角色
const (
	RoleOwner  = 1 // 群主
	RoleAdmin  = 2 // 管理员
	RoleMember = 3 // 普通成员
)

// GroupMemberInfo 群成员信息
type GroupMemberInfo struct {
	UID    int64  `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Role   int    `json:"role"`
}

// GroupInfo 群信息
type GroupInfo struct {
	RoomID int64  `json:"room_id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// CreateGroup 建群（返回群 room_id）
func (s *GroupService) CreateGroup(ownerUID int64, name string, memberUIDs []int64) (int64, error) {
	if name == "" {
		return 0, errors.New("群名称不能为空")
	}

	// ① 建 room（type=1 群聊）
	room := models.Room{Type: 1}
	if err := config.Global.DB.Create(&room).Error; err != nil {
		return 0, err
	}

	// ② 建 room_group（群详情）
	rg := models.RoomGroup{RoomID: int64(room.ID), Name: name}
	if err := config.Global.DB.Create(&rg).Error; err != nil {
		return 0, err
	}

	// ③ 群主加入
	if err := config.Global.DB.Create(&models.GroupMember{GroupID: int64(room.ID), UID: ownerUID, Role: RoleOwner}).Error; err != nil {
		return 0, err
	}

	// ④ 初始成员加入
	for _, uid := range memberUIDs {
		if uid == ownerUID {
			continue
		}
		config.Global.DB.Create(&models.GroupMember{GroupID: int64(room.ID), UID: uid, Role: RoleMember})
	}

	return int64(room.ID), nil
}

// SendGroupMsg 发送群消息（查群成员 → 入管道循环推送）
func (s *GroupService) SendGroupMsg(fromUID, roomID int64, content string, msgType int) (*models.Message, error) {
	// ① 校验是群成员
	var member models.GroupMember
	if err := config.Global.DB.Where("group_id = ? AND uid = ?", roomID, fromUID).First(&member).Error; err != nil {
		return nil, errors.New("你不是该群成员")
	}

	// ② 查所有群成员 uid
	var members []models.GroupMember
	config.Global.DB.Where("group_id = ?", roomID).Find(&members)
	uids := make([]int64, 0, len(members))
	for _, m := range members {
		uids = append(uids, m.UID)
	}

	// ③ 构造消息，入管道（异步推送 + 落库）
	msg := models.Message{
		RoomID:     roomID,
		FromUserID: fromUID,
		Content:    content,
		Type:       msgType,
	}
	msg.CreatedAt = time.Now()
	msgCh <- messageTask{Msg: msg, ToUserIDs: uids}
	return &msg, nil
}

// AddMember 添加群成员（群主/管理员操作）
func (s *GroupService) AddMember(operatorUID, roomID, targetUID int64) error {
	// ① 操作者权限
	var op models.GroupMember
	if err := config.Global.DB.Where("group_id = ? AND uid = ?", roomID, operatorUID).First(&op).Error; err != nil {
		return errors.New("你不是群成员")
	}
	if op.Role != RoleOwner && op.Role != RoleAdmin {
		return errors.New("无权限添加成员")
	}

	// ② 防重复
	var count int64
	config.Global.DB.Model(&models.GroupMember{}).Where("group_id = ? AND uid = ?", roomID, targetUID).Count(&count)
	if count > 0 {
		return errors.New("该用户已在群里")
	}

	// ③ 加入
	return config.Global.DB.Create(&models.GroupMember{GroupID: roomID, UID: targetUID, Role: RoleMember}).Error
}

// RemoveMember 移除群成员（群主踢人，管理员只能踢普通成员）
func (s *GroupService) RemoveMember(operatorUID, roomID, targetUID int64) error {
	var op models.GroupMember
	if err := config.Global.DB.Where("group_id = ? AND uid = ?", roomID, operatorUID).First(&op).Error; err != nil {
		return errors.New("你不是群成员")
	}

	var target models.GroupMember
	if err := config.Global.DB.Where("group_id = ? AND uid = ?", roomID, targetUID).First(&target).Error; err != nil {
		return errors.New("该用户不在群里")
	}

	// 权限规则
	if op.Role == RoleOwner {
		if target.Role == RoleOwner {
			return errors.New("不能移除群主")
		}
	} else if op.Role == RoleAdmin {
		if target.Role != RoleMember {
			return errors.New("管理员只能移除普通成员")
		}
	} else {
		return errors.New("无权限移除成员")
	}

	return config.Global.DB.Where("group_id = ? AND uid = ?", roomID, targetUID).Delete(&models.GroupMember{}).Error
}

// QuitGroup 退出群聊（群主不能直接退群）
func (s *GroupService) QuitGroup(uid, roomID int64) error {
	var member models.GroupMember
	if err := config.Global.DB.Where("group_id = ? AND uid = ?", roomID, uid).First(&member).Error; err != nil {
		return errors.New("你不是群成员")
	}
	if member.Role == RoleOwner {
		return errors.New("群主不能退群，请先转让群主或解散群")
	}
	return config.Global.DB.Where("group_id = ? AND uid = ?", roomID, uid).Delete(&models.GroupMember{}).Error
}

// GetGroupMemberList 获取群成员列表（含用户信息）
func (s *GroupService) GetGroupMemberList(roomID int64) ([]GroupMemberInfo, error) {
	var members []models.GroupMember
	if err := config.Global.DB.Where("group_id = ?", roomID).Find(&members).Error; err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return []GroupMemberInfo{}, nil
	}

	uids := make([]int64, 0, len(members))
	for _, m := range members {
		uids = append(uids, m.UID)
	}
	var users []models.UserBasic
	config.Global.DB.Where("id IN ?", uids).Find(&users)
	userMap := make(map[int64]models.UserBasic, len(users))
	for _, u := range users {
		userMap[int64(u.ID)] = u
	}

	result := make([]GroupMemberInfo, 0, len(members))
	for _, m := range members {
		info := GroupMemberInfo{UID: m.UID, Role: m.Role}
		if u, ok := userMap[m.UID]; ok {
			info.Name = u.Name
			info.Avatar = u.Avatar
		}
		result = append(result, info)
	}
	return result, nil
}

// GetMyGroups 获取我加入的群列表
func (s *GroupService) GetMyGroups(uid int64) ([]GroupInfo, error) {
	var members []models.GroupMember
	if err := config.Global.DB.Where("uid = ?", uid).Find(&members).Error; err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return []GroupInfo{}, nil
	}

	roomIDs := make([]int64, 0, len(members))
	for _, m := range members {
		roomIDs = append(roomIDs, m.GroupID)
	}
	var groups []models.RoomGroup
	config.Global.DB.Where("room_id IN ?", roomIDs).Find(&groups)

	result := make([]GroupInfo, 0, len(groups))
	for _, g := range groups {
		result = append(result, GroupInfo{RoomID: g.RoomID, Name: g.Name, Avatar: g.Avatar})
	}
	return result, nil
}

// UpdateGroup 更新群信息（群名/头像，群主或管理员操作）
func (s *GroupService) UpdateGroup(operatorUID, roomID int64, name, avatar string) error {
	var op models.GroupMember
	if err := config.Global.DB.Where("group_id = ? AND uid = ?", roomID, operatorUID).First(&op).Error; err != nil {
		return errors.New("你不是群成员")
	}
	if op.Role != RoleOwner && op.Role != RoleAdmin {
		return errors.New("无权限修改群信息")
	}

	updates := map[string]any{}
	if name != "" {
		updates["name"] = name
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if len(updates) == 0 {
		return errors.New("没有要更新的字段")
	}
	return config.Global.DB.Model(&models.RoomGroup{}).Where("room_id = ?", roomID).Updates(updates).Error
}
