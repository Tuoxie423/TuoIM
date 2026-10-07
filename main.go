package main

import (
	"log"
	"net/http"

	"ginchat/config"
	"ginchat/initial"
	"ginchat/router"
	"ginchat/service"
	"ginchat/ws"

	_ "ginchat/docs" // swagger 文档注册
)

// @title           GinIM API
// @version         1.0
// @description     即时通讯服务接口文档
// @host            localhost:8080
// @BasePath        /api

func main() {
	config.LoadConfig()
	initial.InitMySQL()
	initial.InitRedis()
	initial.InitLogger()

	// 启动消息队列（异步落库 + 推送）
	service.StartMessageQueue()

	// WebSocket 服务（单独端口，与 HTTP 物理隔离）
	go func() {
		http.HandleFunc("/ws", ws.Connect)
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	r := router.Router()
	r.Run(config.Global.Port.Server)

	service.StopMessageQueue() // 优雅关闭：等管道清空
	initial.Logger.Close()
}
