package api

import (
	"strconv"

	"ginchat/service"
	"ginchat/utils"

	"github.com/gin-gonic/gin"
)

var groupService = new(service.GroupService)

// CreateGroupReq 建群请求
type CreateGroupReq struct {
	Name       string  `json:"name"`
	MemberUIDs []int64 `json:"member_uids"` // 初始成员（可选）
}

// SendGroupMsgReq 发群消息请求
type SendGroupMsgReq struct {
	RoomID  int64  `json:"room_id"`
	Content string `json:"content"`
	Type    int    `json:"type"`
}

// MemberReq 加人/踢人请求
type MemberReq struct {
	RoomID    int64 `json:"room_id"`
	TargetUID int64 `json:"target_uid"`
}

// CreateGroup godoc
// @Summary      创建群聊
// @Tags         群聊
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        body body CreateGroupReq true "建群参数"
// @Success      200 {object} utils.Response
// @Router       /group/create [post]
func CreateGroup(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	var req CreateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	roomID, err := groupService.CreateGroup(uid, req.Name, req.MemberUIDs)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, gin.H{"room_id": roomID})
}

// SendGroupMsg godoc
// @Summary      发送群消息
// @Tags         群聊
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        body body SendGroupMsgReq true "群消息参数"
// @Success      200 {object} utils.Response
// @Router       /group/send [post]
func SendGroupMsg(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	var req SendGroupMsgReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	msg, err := groupService.SendGroupMsg(uid, req.RoomID, req.Content, req.Type)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, msg)
}

// AddGroupMember godoc
// @Summary      添加群成员
// @Tags         群聊
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        body body MemberReq true "成员参数"
// @Success      200 {object} utils.Response
// @Router       /group/add [post]
func AddGroupMember(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	var req MemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	if err := groupService.AddMember(uid, req.RoomID, req.TargetUID); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "已添加")
}

// RemoveGroupMember godoc
// @Summary      移除群成员
// @Tags         群聊
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        body body MemberReq true "成员参数"
// @Success      200 {object} utils.Response
// @Router       /group/remove [post]
func RemoveGroupMember(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	var req MemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	if err := groupService.RemoveMember(uid, req.RoomID, req.TargetUID); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "已移除")
}

// QuitGroup godoc
// @Summary      退出群聊
// @Tags         群聊
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        body body MemberReq true "群会话 room_id"
// @Success      200 {object} utils.Response
// @Router       /group/quit [post]
func QuitGroup(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	var req MemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	if err := groupService.QuitGroup(uid, req.RoomID); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "已退群")
}

// GetGroupMemberList godoc
// @Summary      群成员列表
// @Tags         群聊
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        room_id query integer true "群会话 id"
// @Success      200 {object} utils.Response
// @Router       /group/members [get]
func GetGroupMemberList(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Query("room_id"), 10, 64)
	if roomID == 0 {
		utils.Error(c, 400, "room_id 不能为空")
		return
	}
	list, err := groupService.GetGroupMemberList(roomID)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, list)
}

// GetMyGroups godoc
// @Summary      我的群列表
// @Tags         群聊
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Success      200 {object} utils.Response
// @Router       /group/list [get]
func GetMyGroups(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	list, err := groupService.GetMyGroups(uid)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, list)
}

// UpdateGroupReq 更新群信息请求
type UpdateGroupReq struct {
	RoomID int64  `json:"room_id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// UpdateGroup godoc
// @Summary      更新群信息
// @Tags         群聊
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        body body UpdateGroupReq true "群信息"
// @Success      200 {object} utils.Response
// @Router       /group/update [post]
func UpdateGroup(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	var req UpdateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	if err := groupService.UpdateGroup(uid, req.RoomID, req.Name, req.Avatar); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "群信息已更新")
}
