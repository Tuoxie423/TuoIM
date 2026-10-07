package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ginchat/config"
	"ginchat/initial"
	"ginchat/router"
	"ginchat/service"
	"ginchat/ws"

	_ "ginchat/docs" // swagger 文档注册
)

// @title           TuoIM API
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

	// 监听退出信号（Ctrl+C / kill），触发优雅关闭
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// WebSocket 服务（单独端口，与 HTTP 物理隔离）
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/ws", ws.Connect)
	wsServer := &http.Server{Addr: ":8081", Handler: wsMux}
	go func() {
		if err := wsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[main] WebSocket 服务启动失败: %v", err)
		}
	}()

	// HTTP API 服务
	apiServer := &http.Server{Addr: config.Global.Port.Server, Handler: router.Router()}
	go func() {
		if err := apiServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[main] HTTP 服务启动失败: %v", err)
		}
	}()

	log.Printf("[main] HTTP 监听 %s，WebSocket 监听 :8081", config.Global.Port.Server)

	// 阻塞等待退出信号
	<-ctx.Done()
	log.Println("[main] 收到退出信号，开始优雅关闭...")
	stop() // 停止捕获信号，之后再次 Ctrl+C 可强制退出

	// ① 先停 HTTP / WebSocket：停止接收新请求，等在途请求完成（最多 5s）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("[main] HTTP 优雅关闭失败: %v", err)
	}
	if err := wsServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("[main] WebSocket 优雅关闭失败: %v", err)
	}

	// ② 再停消息队列：清空管道，刷掉攒批未落库的消息
	service.StopMessageQueue()

	// ③ 最后关闭日志：刷盘
	initial.Logger.Close()
	log.Println("[main] 已退出")
}
