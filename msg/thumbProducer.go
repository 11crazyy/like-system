package msg

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"index/models"
	"index/models/enum"
	"index/util"
	"strconv"
	"time"
)

func DoThumb(ctx *gin.Context, db *gorm.DB, blogId int64, redis *redis.Client, producer pulsar.Producer) (bool, error) {
	loginUser := models.CurrentUser(ctx, db)
	loginUserId := loginUser.ID
	userThumbKey := util.GetUserThumbKey(loginUserId)
	//执行redis Lua脚本
	redisLuaScript := models.GetNewThumbScript()
	res, err := redisLuaScript.Run(ctx, redis, []string{userThumbKey}, blogId, strconv.Itoa(int(loginUserId))).Int()
	if err != nil {
		return false, err
	}
	if enum.LuaStatus(res) == enum.LuaFail {
		return false, errors.New("用户已经点赞")
	}
	thumbEvent := ThumbEvent{
		BlogId:    blogId,
		Type:      EventTypeINCR,
		UserId:    int64(loginUserId),
		EventTime: time.Now(),
	}
	sendAsyncWithFallback(producer, &thumbEvent, func() {
		//如果消息队列发送失败，则redis回滚 删除userThumbKey blogId的记录
		redis.HDel(ctx, userThumbKey, strconv.Itoa(int(blogId)))
	})
	return true, nil
}

// 发送消息队列消息实现数据的持久化 使得redis只负责记录用户点赞信息的更新，持久化交给 消息队列，进一步解藕
func sendAsyncWithFallback(producer pulsar.Producer, event *ThumbEvent, fallback func()) {
	payload, _ := json.Marshal(event)
	producer.SendAsync(context.Background(), &pulsar.ProducerMessage{
		Payload: payload,
	}, func(msgID pulsar.MessageID, message *pulsar.ProducerMessage, err error) {
		if err != nil {
			logrus.Error("点赞事件发送失败", err)
			fallback()
		}
	})
}

func UndoThumb(ctx *gin.Context, db *gorm.DB, blogId int64, redis *redis.Client, producer pulsar.Producer) (bool, error) {
	loginUser := models.CurrentUser(ctx, db)
	loginUserId := loginUser.ID
	userThumbKey := util.GetUserThumbKey(loginUserId)
	//执行lua脚本取消点赞
	redisLuaScript := models.GetNewUnthumbScript()
	res, err := redisLuaScript.Run(ctx, redis, []string{userThumbKey}, blogId, strconv.Itoa(int(loginUserId))).Int()
	if err != nil {
		return false, err
	}
	if enum.LuaStatus(res) == enum.LuaFail {
		return false, errors.New("用户没有点赞")
	}
	// 发送消息队列消息
	thumbEvent := ThumbEvent{
		BlogId:    blogId,
		Type:      EventTypeDECR,
		UserId:    int64(loginUserId),
		EventTime: time.Now(),
	}
	sendAsyncWithFallback(producer, &thumbEvent, func() {
		//如果消息队列发送失败，则redis回滚 增加userThumbKey blogId的记录
		redis.HSet(ctx, userThumbKey, strconv.Itoa(int(blogId)), 1)
	})
	return true, nil
}
