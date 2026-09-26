package initial

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"ginchat/config"
	"ginchat/middleware"
)

// Logger 全局异步日志器单例
var Logger *middleware.AsyncLogger

// InitLogger 按 config.Log 配置初始化异步日志器
func InitLogger() {
	var w io.Writer
	switch config.Global.Log.Output {
	case "file", "both":
		// 先确保日志目录存在（O_CREATE 只建文件不建目录）
		if err := os.MkdirAll(filepath.Dir(config.Global.Log.File), 0755); err != nil {
			log.Fatalf("Failed to create log dir: %v", err)
		}
		f, err := os.OpenFile(config.Global.Log.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		if config.Global.Log.Output == "both" {
			w = io.MultiWriter(os.Stdout, f) // 终端 + 文件都写
		} else {
			w = f
		}
	default:
		w = os.Stdout
	}

	// TODO: 若要支持 json 格式，可在 config.Log 里加回 Format 字段后再传 true
	Logger = middleware.NewAsyncLogger(w, 1024, config.Global.Log.FormatJson)
}
