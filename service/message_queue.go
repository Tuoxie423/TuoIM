package service

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"ginchat/cache"
	"ginchat/config"
	"ginchat/models"
	"ginchat/ws"
)

// messageTask 消息任务（消息 + 接收者）
type messageTask struct {
	Msg      models.Message
	ToUserID int64
}

var (
	msgCh     = make(chan messageTask, 1024)    // 消息主管道
	persistCh = make(chan models.Message, 1024) // 落库管道
	wg        sync.WaitGroup
)

// StartMessageQueue 启动消息队列（main.go 调用）
func StartMessageQueue() {
	wg.Add(3)
	go dispatcher()
	go persister()
	go retrier()
}

// StopMessageQueue 优雅关闭：停止接收新消息，等管道清空
func StopMessageQueue() {
	close(msgCh)
	wg.Wait()
}

// dispatcher 分发器：推送给接收方 → 丢进落库管道
func dispatcher() {
	defer wg.Done()
	for task := range msgCh {
		msgJSON, _ := json.Marshal(task.Msg)
		if cache.IsOnline(task.ToUserID) {
			log.Printf("[MQ] 在线推送: from=%d to=%d", task.Msg.FromUserID, task.ToUserID)
			ws.DefaultHub.Push(task.ToUserID, msgJSON) // 在线：WebSocket 推
		} else {
			log.Printf("[MQ] 离线存队列: from=%d to=%d", task.Msg.FromUserID, task.ToUserID)
			cache.SaveOfflineMsg(task.ToUserID, msgJSON) // 不在线：存离线队列
		}
		persistCh <- task.Msg
	}
	close(persistCh)
}

// persister 落库器：攒批批量落库，失败进死信队列
func persister() {
	defer wg.Done()
	batch := make([]models.Message, 0, 100)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-persistCh:
			if !ok {
				flushBatch(batch) // 管道关闭，刷掉剩余
				return
			}
			batch = append(batch, msg)
			if len(batch) >= 100 {
				flushBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 { // 到点没攒满也刷
				flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch 批量落库，失败则进死信队列
func flushBatch(batch []models.Message) {
	if err := config.Global.DB.CreateInBatches(batch, 100).Error; err != nil {
		log.Printf("[MQ] 落库失败，进死信队列: err=%v, 条数=%d", err, len(batch))
		for _, msg := range batch {
			msgJSON, _ := json.Marshal(msg)
			cache.SaveFailMsg(msgJSON)
		}
	}
}

// retrier 重试器：定时重试死信队列里的消息
func retrier() {
	defer wg.Done()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, data := range cache.GetFailMsg() {
			var msg models.Message
			if err := json.Unmarshal([]byte(data), &msg); err != nil {
				log.Printf("[MQ] 死信消息反序列化失败: err=%v", err)
				continue
			}
			if err := config.Global.DB.Create(&msg).Error; err == nil {
				cache.RemoveFailMsg(data) // 成功 → 删除
			} else {
				log.Printf("[MQ] 死信重试失败: err=%v", err)
				// 失败 → 留着，下次再试
			}
		}
	}
}
