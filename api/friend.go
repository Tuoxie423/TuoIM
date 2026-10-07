package api

import (
	"ginchat/service"
	"ginchat/utils"

	"github.com/gin-gonic/gin"
)

var friendService = new(service.FriendService)

// getCurrentUID 从 context 取当前登录用户 id（JWTAuth 中间件存入的 claims）
func getCurrentUID(c *gin.Context) (int64, bool) {
	claimsAny, ok := c.Get("claims")
	if !ok {
		return 0, false
	}
	claims, ok := claimsAny.(*utils.Claims)
	if !ok {
		return 0, false
	}
	return claims.UserID, true
}

// ApplyFriendReq 发起好友申请请求
type ApplyFriendReq struct {
	ToUserID int64  `json:"to_user_id"`
	Msg      string `json:"msg"`
}

// AgreeReq 同意/拒绝申请请求
type AgreeReq struct {
	ApplyUID int64 `json:"apply_uid"` // 申请人 uid
}

// DeleteFriendReq 删除好友请求
type DeleteFriendReq struct {
	FriendUID int64 `json:"friend_uid"` // 要删除的好友 uid
}

// SearchUser godoc
// @Summary      搜索用户
// @Tags         好友
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        phone query string true "手机号"
// @Success      200 {object} utils.Response
// @Router       /users/search [get]
func SearchUser(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		utils.Error(c, 400, "手机号不能为空")
		return
	}
	user, err := friendService.SearchUser(phone)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, user)
}

// ApplyFriend godoc
// @Summary      发起好友申请
// @Tags         好友
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        body body ApplyFriendReq true "申请参数"
// @Success      200 {object} utils.Response
// @Router       /friend/apply [post]
func ApplyFriend(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	var req ApplyFriendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	if err := friendService.ApplyFriend(uid, req.ToUserID, req.Msg); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "好友申请已发送")
}

// AgreeFriend godoc
// @Summary      同意好友申请
// @Tags         好友
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        body body AgreeReq true "申请人 uid"
// @Success      200 {object} utils.Response
// @Router       /friend/agree [post]
func AgreeFriend(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	var req AgreeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	if err := friendService.AgreeFriend(uid, req.ApplyUID); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}

// RejectFriend godoc
// @Summary      拒绝好友申请
// @Tags         好友
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        body body AgreeReq true "申请人 uid"
// @Success      200 {object} utils.Response
// @Router       /friend/reject [post]
func RejectFriend(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	var req AgreeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	if err := friendService.RejectFriend(uid, req.ApplyUID); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}

// GetApplyList godoc
// @Summary      好友申请列表
// @Tags         好友
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Success      200 {object} utils.Response
// @Router       /friend/applyList [get]
func GetApplyList(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	list, err := friendService.GetApplyList(uid)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, list)
}

// GetFriendList godoc
// @Summary      好友列表
// @Tags         好友
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Success      200 {object} utils.Response
// @Router       /friend/list [get]
func GetFriendList(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	list, err := friendService.GetFriendList(uid)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, list)
}

// UnreadApplyNum godoc
// @Summary      未读申请数
// @Tags         好友
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Success      200 {object} utils.Response
// @Router       /friend/unreadNum [get]
func UnreadApplyNum(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	num, err := friendService.UnreadApplyNum(uid)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, num)
}

// DeleteFriend godoc
// @Summary      删除好友
// @Tags         好友
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        body body DeleteFriendReq true "好友 uid"
// @Success      200 {object} utils.Response
// @Router       /friend/delete [delete]
func DeleteFriend(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	var req DeleteFriendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	if err := friendService.DeleteFriend(uid, req.FriendUID); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}
