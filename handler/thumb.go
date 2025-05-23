package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"gorm.io/gorm"
	"index/cache"
	"index/models"
	"net/http"
	"strconv"
	"sync"
)

func (h *Handlers) handleDoThumb(c *gin.Context) {
	blogId := c.Param("blogId")
	if blogId == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, errors.New("参数不能为空"))
		return
	}
	user := models.CurrentUser(c, h.db)
	if user == nil {
		carrot.AbortWithJSONError(c, http.StatusUnauthorized, errors.New("用户未登录"))
		return
	}
	mutex := sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()

	//在事务中执行点赞操作
	err := h.db.Transaction(func(tx *gorm.DB) error {
		//err, count := models.GetBlogByUserIdAndBlogId(tx, blogId, strconv.Itoa(int(user.ID)))
		//if err != nil {
		//	return err
		//}
		//if count != 0 {
		//	return errors.New("用户已经点赞")
		//}
		bId, _ := strconv.ParseInt(blogId, 10, 64)
		exist, err := cache.HasThumb(bId, int64(user.ID), h.redis)
		if err != nil {
			return err
		}
		if exist {
			return errors.New("用户已经点赞")
		}

		blogID, _ := strconv.ParseInt(blogId, 10, 64)
		blog, err := models.GetBlogById(tx, blogID)
		if err != nil {
			return err
		}
		//更新博客点赞数
		if err := models.UpdateThumbNum(tx, blogId, blog.ThumbCount+1); err != nil {
			return err
		}

		thumb := models.Thumb{
			UserID: int64(user.ID),
			BlogID: blogID,
		}
		if err := models.CreateThumb(tx, &thumb); err != nil {
			return err
		}
		//将点赞记录存入redis
		err = cache.SavaThumb(blogID, int64(user.ID), h.redis)
		return err
	})
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "点赞成功",
	})
}

func (h *Handlers) handleCancelThumb(c *gin.Context) {
	//取消点赞接口 1.判断参数 2.获取当前用户 3.加锁 事务 判断当前用户是否点赞该博客 4.有则更新数据库该blog的点赞数，-1，同时数据库中删除这条thumb记录
	blogId := c.Param("blogId")
	if blogId == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, errors.New("参数不能为空"))
		return
	}
	user := models.CurrentUser(c, h.db)
	if user == nil {
		carrot.AbortWithJSONError(c, http.StatusUnauthorized, errors.New("用户未登录"))
		return
	}
	mutex := sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()

	err := h.db.Transaction(func(tx *gorm.DB) error {
		//err, count := models.GetBlogByUserIdAndBlogId(tx, blogId, strconv.Itoa(int(user.ID)))
		//if err != nil {
		//	return err
		//}
		//if count == 0 {
		//	return errors.New("用户没有点赞")
		//}
		bId, _ := strconv.ParseInt(blogId, 10, 64)
		exist, err := cache.HasThumb(bId, int64(user.ID), h.redis)
		if err != nil {
			return err
		}
		if !exist {
			return errors.New("用户没有点赞")
		}

		blogID, _ := strconv.ParseInt(blogId, 10, 64)
		blog, err := models.GetBlogById(tx, blogID)
		if err != nil {
			return err
		}
		//更新博客点赞数
		if err := models.UpdateThumbNum(tx, blogId, blog.ThumbCount-1); err != nil {
			return err
		}
		if err := models.DeleteThumb(tx, blogId, strconv.Itoa(int(user.ID))); err != nil {
			return err
		}
		if err := cache.DeleteThumb(blogID, int64(user.ID), h.redis); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "取消点赞成功",
	})
}
