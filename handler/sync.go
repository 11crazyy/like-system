package handler

import (
	"context"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"index/models"
	"strconv"
	"strings"
	"time"
)

func (h *Handlers) Start() {
	ticker := time.NewTicker(10 * time.Second) //每十秒同步一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.Run()
		}
	}
}

func (h *Handlers) Run() {
	logrus.Info("开始执行同步任务")
	timeSlice := models.GetPreviousTimeSlice()
	h.syncThumbToDbByData(timeSlice)
}

func (h *Handlers) syncThumbToDbByData(timeSlice string) {
	//同步上一个时间片的数据到数据库
	ctx := context.Background()
	tempThumbKey := models.TEMP_THUMB_KEY_PREFIX + timeSlice

	//从redis中获取tempThumbKey对应的所有临时点赞数据
	allTempThumb, err := h.redis.HGetAll(ctx, tempThumbKey).Result()
	if err != nil {
		logrus.Error(err)
		return
	}
	if len(allTempThumb) == 0 {
		return
	}

	//处理每条数据
	toAdd := make([]*models.Thumb, 0)
	toDelete := make([]*models.Thumb, 0)
	countMap := make(map[int64]int64)
	for userBlogId, thumb := range allTempThumb {
		parts := strings.Split(userBlogId, ":")
		userId, _ := strconv.ParseUint(parts[0], 10, 64)
		blogId, _ := strconv.ParseUint(parts[1], 10, 64)
		opType, _ := strconv.Atoi(thumb)

		switch opType {
		case 1: //点赞
			toAdd = append(toAdd, &models.Thumb{
				UserID: int64(userId),
				BlogID: int64(blogId),
			})
		case -1: //取消点赞
			toDelete = append(toDelete, &models.Thumb{
				UserID: int64(userId),
				BlogID: int64(blogId),
			})
		default:
			logrus.Error("无效的操作类型")
			continue
		}
		countMap[int64(blogId)] += int64(opType)
	}

	//在事务中执行数据库操作
	err = h.db.Transaction(func(tx *gorm.DB) error {
		//批量创建点赞数据
		if len(toAdd) > 0 {
			err = models.BatchCreateThumb(h.db, toAdd)
			if err != nil {
				return err
			}
		}
		//批量删除取消点赞的记录
		if len(toDelete) > 0 {
			for _, op := range toDelete {
				err := models.DeleteThumb(h.db, strconv.FormatInt(op.BlogID, 10), strconv.FormatInt(op.UserID, 10))
				if err != nil {
					return err
				}
			}
		}
		//批量更新blog的点赞数
		if len(countMap) > 0 {
			err := models.BatchUpdateThumbCount(h.db, countMap)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logrus.WithField("err", "数据同步到数据库失败").Error(err)
		return
	}

	//异步删除Redis中处理过的数据
	go func() {
		if _, err := h.redis.Del(ctx, tempThumbKey).Result(); err != nil {
			logrus.WithField("err", "删除Redis键失败").Error(err)
		}
	}()
}

func (h *Handlers) StartCompensatoryJob() {
	now := time.Now()
	todayToAM := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	var next2AM time.Time
	if now.After(todayToAM) {
		next2AM = todayToAM.AddDate(0, 0, 1)
	} else {
		next2AM = todayToAM
	}
	duration := next2AM.Sub(now)
	time.Sleep(duration)

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.SyncThumbToDbCompensatory()
		}
	}
}

// SyncThumbToDbCompensatory 补偿机制 同步未同步完的数据
func (h *Handlers) SyncThumbToDbCompensatory() {
	logrus.Info("开始执行补偿机制")
	pattern := models.TEMP_THUMB_KEY_PREFIX + "*"
	thumbKeys, err := h.redis.Keys(context.Background(), pattern).Result()
	if err != nil {
		logrus.WithField("redisKey", thumbKeys).Error(err)
		return
	}

	//提取需要补偿的时间片 按时间片顺序重新执行同步
	timeSlices := make(map[string]bool)
	for _, key := range thumbKeys {
		//从thumb:temp:14:25:30中提取14:25:30
		parts := strings.Split(key, ":")
		if len(parts) >= 3 {
			timeSlice := strings.Join(parts[2:], ":")
			timeSlices[timeSlice] = true
		}
	}
	for timeSlice := range timeSlices {
		logrus.Info("开始执行补偿机制任务,时间片：", timeSlice)
		h.syncThumbToDbByData(timeSlice)
	}
	logrus.Info("补偿机制执行完毕")
}
