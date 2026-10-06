package api

import (
	"ginchat/service"
	"ginchat/utils"

	"github.com/gin-gonic/gin"
)

var messageService = new(service.MessageService)

// MessageReq 发送消息请求
type MessageReq struct {
	ToUserID int64  `json:"to_user_id"` // 接收者用户 id
	Content  string `json:"content"`    // 消息内容
	Type     int    `json:"type"`       // 消息类型 1文本 2图片...
}

// SendMessage godoc
// @Summary      发送消息
// @Tags         消息
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        body body MessageReq true "消息参数"
// @Success      200 {object} utils.Response
// @Router       /chat/send [post]
func SendMessage(c *gin.Context) {
	// 从 token 解析出发送者 id
	claimsAny, ok := c.Get("claims")
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	userClaims, ok := claimsAny.(*utils.Claims)
	if !ok {
		utils.Error(c, 500, "凭证解析失败")
		return
	}

	var req MessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	msg, err := messageService.SendMsg(userClaims.UserID, req.ToUserID, req.Content, req.Type)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, msg)
}

func GetRoomID(c *gin.Context) {

}
