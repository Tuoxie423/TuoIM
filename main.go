package main

import (
	"ginchat/config"
	"ginchat/initial"
	"ginchat/router"

	_ "ginchat/docs" // swagger 文档注册
)

// @title           GinIM API
// @version         1.0
// @description     即时通讯服务接口文档
// @host            localhost:8080
// @BasePath        /

func main() {
	config.LoadConfig()
	initial.InitMySQL()
	initial.InitRedis()
	initial.InitLogger()

	r := router.Router()
	// 启动 websocket 连接中心（必须跑在 goroutine 里）
	// go ws.DefaultHub.Run()
	r.Run(config.Global.Port.Server)
	initial.Logger.Close()
}
