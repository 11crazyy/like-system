package msg

import (
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

// 只对帐有改动的数据其他的不管 多线程对帐 定时任务 每天2点执行
func ReconcileJob(db *gorm.DB, redis *redis.Client, pulsarURL, topic, subName string, batchSize, workers int, flushInterval time.Duration) {
	//获取时间分片下的所有用户id
	
	//对于每个userId 比较redis和mysql的数据差异

	//计算差异（redis有但mysql无）

	//发送补偿事件到pulsar
}

func sendCompensationEvents(userId int64, blogIds []int64, producer pulsar.Producer) {
	for _, blogId := range blogIds {
		event := &ThumbEvent{
			UserId:    userId,
			BlogId:    blogId,
			Type:      EventTypeINCR,
			EventTime: time.Now(),
		}
		sendAsyncWithFallback(producer, event, func() {
			logrus.Error("补偿事件发送失败")
		})
	}
}
