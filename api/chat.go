package api

import (
	"strconv"

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
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}

	var req MessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	msg, err := messageService.SendMsg(uid, req.ToUserID, req.Content, req.Type)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, msg)
}

// GetHistory godoc
// @Summary      消息历史
// @Tags         消息
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        room_id query string true "会话 id"
// @Param        cursor query string false "游标（上一页最后一条消息 id）"
// @Param        limit query string false "条数，默认20"
// @Success      200 {object} utils.Response
// @Router       /chat/history [get]
func GetHistory(c *gin.Context) {
	uid, ok := getCurrentUID(c)
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	roomID, _ := strconv.ParseInt(c.Query("room_id"), 10, 64)
	cursor, _ := strconv.ParseInt(c.Query("cursor"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if roomID == 0 {
		utils.Error(c, 400, "room_id 不能为空")
		return
	}

	list, hasMore, err := messageService.GetHistory(uid, roomID, cursor, limit)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{"list": list, "has_more": hasMore})
}
