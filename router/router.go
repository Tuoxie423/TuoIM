package router

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"ginchat/api"
	"ginchat/initial"
	"ginchat/middleware"
)

func Router() *gin.Engine {
	r := gin.New()
	gin.SetMode(gin.DebugMode)
	r.Use(initial.Logger.Logger()) // 使用自定义日志中间件
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	// 公开接口
	publicGroup := r.Group("/api")
	{
		// swagger 文档
		publicGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

		// 首页
		publicGroup.GET("/index", api.Index)

		// 认证
		publicGroup.POST("/login", api.Login)
		publicGroup.POST("/register", api.Register)
	}

	// 受保护接口（需要 JWT）
	users := publicGroup.Group("/users")
	users.Use(middleware.JWTAuth())
	{
		users.GET("/info", api.GetUserInfo)
		users.PUT("/info", api.UpdateUserInfo)
		users.GET("/search", api.SearchUser)
	}

	// 好友（受保护）
	friend := publicGroup.Group("/friend")
	friend.Use(middleware.JWTAuth())
	{
		friend.POST("/apply", api.ApplyFriend)
		friend.POST("/agree", api.AgreeFriend)
		friend.POST("/reject", api.RejectFriend)
		friend.DELETE("/delete", api.DeleteFriend)
		friend.GET("/applyList", api.GetApplyList)
		friend.GET("/list", api.GetFriendList)
		friend.GET("/unreadNum", api.UnreadApplyNum)
	}

	// 消息（受保护）
	chat := publicGroup.Group("/chat")
	chat.Use(middleware.JWTAuth())
	{
		chat.POST("/send", api.SendMessage)
		chat.GET("/history", api.GetHistory)
	}

	// 房间（受保护）
	room := publicGroup.Group("/room")
	room.Use(middleware.JWTAuth())
	{
		room.POST("/get_or_create", api.GetOrCreateRoom)
	}

	return r
}
