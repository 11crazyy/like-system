package msg

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/sirupsen/logrus"
	"time"
)

type BatchConsumerConfig struct {
	MaxNumMessage int           //每次最大处理消息数
	Timeout       time.Duration //批处理超时时间
}

// 创建带批量消费的消费者
func NewBatchConsumer(client pulsar.Client, topic string, subscription string) (pulsar.Consumer, error) {
	consumerOptions := pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: subscription,
		Type:             pulsar.Shared, //共享订阅模式
	}
	//创建消费者
	consumer, err := client.Subscribe(consumerOptions)
	if err != nil {
		return nil, err
	}
	return consumer, nil
}

// 批量消费信息
func BatchConsume(consumer pulsar.Consumer, config BatchConsumerConfig, handler func([]pulsar.Message) error) {
	ctx := context.Background()
	var batch []pulsar.Message
	timer := time.NewTimer(config.Timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			//	超时处理当前批次
			if len(batch) > 0 {
				if err := handler(batch); err != nil {
					logrus.Error("处理批次失败：%v", err)
				}
				batch = nil
			}
			timer.Reset(config.Timeout)
		default:
			//接收单条消息
			msg, err := consumer.Receive(ctx)
			if err != nil {
				logrus.Error("接收消息失败:%v", err)
			}
			batch = append(batch, msg)

			if len(batch) >= config.MaxNumMessage {
				if err := handler(batch); err != nil {
					logrus.Error("处理批次失败：%v", err)
				}
				batch = nil
				timer.Reset(config.Timeout)
			}
		}

	}

}
