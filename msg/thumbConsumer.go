package msg

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"index/models"

	_ "log"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"gorm.io/gorm"
)

type ThumbConsumer struct {
	db            *gorm.DB
	client        pulsar.Client
	consumer      pulsar.Consumer
	batchSize     int
	flushInterval time.Duration
	msgChan       chan pulsar.Message
	workerCount   int
	wg            sync.WaitGroup
	shutdown      chan struct{}
}

func NewThumbConsumer(db *gorm.DB, pulsarURL, topic, subName string, batchSize, workers int, flushInterval time.Duration) (*ThumbConsumer, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               pulsarURL,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: subName,
		Type:             pulsar.Shared,
	})
	if err != nil {
		client.Close()
		return nil, err
	}

	return &ThumbConsumer{
		db:            db,
		client:        client,
		consumer:      consumer,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		msgChan:       make(chan pulsar.Message, batchSize*2),
		workerCount:   workers,
		shutdown:      make(chan struct{}),
	}, nil
}

func (c *ThumbConsumer) Start() {
	c.wg.Add(1)
	go c.receiveMessages()

	for i := 0; i < c.workerCount; i++ {
		c.wg.Add(1)
		go c.processBatches()
	}
}

func (c *ThumbConsumer) receiveMessages() {
	defer c.wg.Done()
	for {
		select {
		case <-c.shutdown:
			return
		default:
			msg, err := c.consumer.Receive(context.Background())
			if err != nil {
				logrus.Error("消息接收失败")
				continue
			}
			c.msgChan <- msg
		}
	}
}

func (c *ThumbConsumer) processBatches() {
	defer c.wg.Done()

	var (
		batch      []pulsar.Message
		flushTimer = time.NewTimer(c.flushInterval)
	)
	defer flushTimer.Stop()

	for {
		select {
		case msg, ok := <-c.msgChan:
			if !ok {
				if len(batch) > 0 {
					c.processBatch(batch)
				}
				return
			}
			batch = append(batch, msg)
			if len(batch) >= c.batchSize {
				c.flushBatch(&batch, flushTimer) // 将消息批量处理
			}
		case <-flushTimer.C:
			c.flushBatch(&batch, flushTimer)
		case <-c.shutdown:
			if len(batch) > 0 {
				c.processBatch(batch)
			}
			return
		}

	}
}

func (c *ThumbConsumer) flushBatch(batch *[]pulsar.Message, timer *time.Timer) {
	if len(*batch) > 0 {
		c.processBatch(*batch)
		*batch = (*batch)[:0] //清空batch
	}
	timer.Reset(c.flushInterval) // 重置定时器
}

func (c *ThumbConsumer) processBatch(messages []pulsar.Message) {
	//	解析消息
	events := make([]*ThumbEvent, 0, len(messages))
	for _, msgg := range messages {
		var event ThumbEvent
		if err := json.Unmarshal(msgg.Payload(), &event); err != nil {
			logrus.Error("消息解析失败")
			c.consumer.Nack(msgg) // 消息解析失败，拒绝消息
			continue
		}
		events = append(events, &event)
	}
	//处理消息
	if err := c.handleEvents(events); err != nil {
		logrus.Error("事件处理失败")
		for _, msg := range messages {
			c.consumer.Nack(msg) // 处理失败，拒绝消息
		}
		return
	}

	// 确认消息
	for _, msg := range messages {
		c.consumer.Ack(msg)
	}
}

func (c *ThumbConsumer) handleEvents(events []*ThumbEvent) interface{} {
	//分组获取最新事件
	latestEvents := make(map[struct{ UserID, BlogID int64 }]*ThumbEvent)
	for _, event := range events {
		key := struct{ UserID, BlogID int64 }{UserID: event.UserId, BlogID: event.BlogId}
		if latestEvent, ok := latestEvents[key]; !ok || event.EventTime.After(latestEvent.EventTime) {
			latestEvents[key] = event
		}
	}

	// 批量处理事件到数据库
	var (
		countMap    = make(map[int64]int64)
		thumbsToAdd = make([]*models.Thumb, 0)
		thumbsToDel = make([]struct{ UserID, BlogID int64 }, 0)
	)

	for key, event := range latestEvents {
		switch event.Type {
		case EventTypeINCR:
			countMap[event.BlogId]++
			thumbsToAdd = append(thumbsToAdd, &models.Thumb{
				UserID: event.UserId,
				BlogID: event.BlogId,
			})
		case EventTypeDECR:
			countMap[event.BlogId]--
			thumbsToDel = append(thumbsToDel, key)
		}
	}

	//执行数据库事务
	return c.db.Transaction(func(tx *gorm.DB) error {
		//添加点赞
		if len(thumbsToAdd) > 0 {
			return tx.CreateInBatches(thumbsToAdd, c.batchSize).Error
		}
		
		//删除取消的点赞
		if len(thumbsToDel) > 0 {
			for _, del := range thumbsToDel {
				if err := tx.Where("user_id = ? AND blog_id = ?", del.UserID, del.BlogID).Delete(&models.Thumb{}).Error; err != nil {
					return err
				}
			}
		}

		//更新博客点赞数
		for blogId, count := range countMap {
			//更新该博客的点赞数
			if err := tx.Model(&models.Blog{}).Where("id = ?", blogId).Update("thumb_count", gorm.Expr("thumb_count + ?", count)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
