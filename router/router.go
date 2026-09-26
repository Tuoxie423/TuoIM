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
	gin.SetMode(gin.ReleaseMode)
	r.Use(initial.Logger.Logger()) // 使用自定义日志中间件
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 首页
	r.GET("/index", api.Index)
	r.GET("/getuser", api.GetUser)

	// 用户
	user := r.Group("/user")
	user.User(middleware.adminGroup)
	{
		user.POST("/register", api.Register)
		user.POST("/login", api.Login)
	}

	return r
}
