package middleware

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// logEntry 一条日志记录
type logEntry struct {
	Time   time.Time
	Level  string
	Status int
	Method string
	Path   string
	IP     string
	Cost   time.Duration
}

// AsyncLogger 异步日志器：业务 goroutine 投递日志，后台 goroutine 写盘
type AsyncLogger struct {
	ch      chan logEntry // 缓冲 channel，缓冲大小决定能扛多大突发量
	out     *bufio.Writer // 缓冲写，攒批刷盘
	useJSON bool          // 是否输出结构化 JSON

	wg       sync.WaitGroup
	once     sync.Once // 保证 Close 只执行一次
	dropped  uint64    // channel 满被丢弃的日志数
	writeErr uint64    // 写盘错误次数
}

// NewAsyncLogger 创建异步日志器并启动后台写 goroutine
// useJSON 为 true 时输出结构化 JSON，否则输出可读文本
func NewAsyncLogger(w io.Writer, size int, useJSON bool) *AsyncLogger {
	if size <= 0 {
		size = 1024 // 兜底，避免无缓冲 channel 导致日志全被丢弃
	}
	l := &AsyncLogger{
		ch:      make(chan logEntry, size),
		out:     bufio.NewWriter(w),
		useJSON: useJSON,
	}
	l.wg.Add(1)
	go l.run()
	return l
}

// write 非阻塞投递一条日志；channel 满时丢弃并计数，绝不阻塞业务
func (l *AsyncLogger) write(e logEntry) {
	select {
	case l.ch <- e:
	default:
		atomic.AddUint64(&l.dropped, 1)
	}
}

// run 后台消费者：写日志 + 定时刷盘
func (l *AsyncLogger) run() {
	defer l.wg.Done()

	// 每 3 秒强制刷一次盘，避免低流量时日志一直堆在内存缓冲里
	flushTicker := time.NewTicker(3 * time.Second)
	defer flushTicker.Stop()

	for {
		select {
		case en, ok := <-l.ch:
			if !ok { // channel 已关闭且缓冲已读完，退出
				return
			}
			if _, err := l.out.WriteString(l.format(en) + "\n"); err != nil {
				atomic.AddUint64(&l.writeErr, 1)
			}
		case <-flushTicker.C:
			if err := l.out.Flush(); err != nil {
				atomic.AddUint64(&l.writeErr, 1)
			}
		}
	}
}

// format 按配置格式化一条日志
func (l *AsyncLogger) format(e logEntry) string {
	if l.useJSON {
		// 用 struct 而非 map，保证 JSON 字段顺序稳定，利于日志系统索引
		b, _ := json.Marshal(struct {
			Time   string `json:"time"`
			Level  string `json:"level"`
			Method string `json:"method"`
			Path   string `json:"path"`
			Status int    `json:"status"`
			CostMs int64  `json:"cost_ms"`
			IP     string `json:"ip"`
		}{
			Time:   e.Time.Format(time.RFC3339),
			Level:  e.Level,
			Method: e.Method,
			Path:   e.Path,
			Status: e.Status,
			CostMs: e.Cost.Milliseconds(),
			IP:     e.IP,
		})
		return string(b)
	}

	// 文本，人眼可读
	return fmt.Sprintf("%s [%s] %s %s %d %v",
		e.Time.Format("2006-01-02 15:04:05"),
		e.Level, e.Method, e.Path, e.Status, e.Cost)
}

// Middleware 记录每个请求的 gin 日志中间件
func (l *AsyncLogger) Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		l.write(logEntry{
			Time:   start,
			Level:  levelFromStatus(c.Writer.Status()),
			Status: c.Writer.Status(),
			Method: c.Request.Method,
			Path:   c.Request.URL.Path,
			IP:     c.ClientIP(),
			Cost:   time.Since(start),
		})
	}
}

// Close 优雅关闭：停止接收 → 等待后台写完 → 刷盘。可安全重复调用
func (l *AsyncLogger) Close() {
	l.once.Do(func() {
		close(l.ch)
		l.wg.Wait()
		_ = l.out.Flush()
	})
}

// Dropped 返回被丢弃的日志条数（channel 满时）
func (l *AsyncLogger) Dropped() uint64 {
	return atomic.LoadUint64(&l.dropped)
}

// WriteErrors 返回写盘错误次数
func (l *AsyncLogger) WriteErrors() uint64 {
	return atomic.LoadUint64(&l.writeErr)
}

// levelFromStatus 按 HTTP 状态码推断日志级别
func levelFromStatus(status int) string {
	switch {
	case status >= 500:
		return "error"
	case status >= 400:
		return "warn"
	default:
		return "info"
	}
}
