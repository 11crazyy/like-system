package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"index/cache"
	"index/models"
	"net/http"
	"strconv"
)

func (h *Handlers) handleGetBlogById(ctx *gin.Context) {
	blogId := ctx.Param("blogId")
	if blogId == "" {
		carrot.AbortWithJSONError(ctx, http.StatusBadRequest, errors.New("blogId不能为空"))
		return
	}
	loginUser := models.CurrentUser(ctx, h.db)

	blogID, _ := strconv.ParseInt(blogId, 10, 64)
	blog, _ := models.GetBlogById(h.db, blogID)

	thumb, err := cache.HasThumb(blogID, int64(loginUser.ID), h.redis)
	if err != nil {
		carrot.AbortWithJSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	//给blog填充是否被点赞的属性
	blog.HasThumb = thumb
	ctx.JSON(http.StatusOK, blog)
}

func (h *Handlers) GetBlogList(ctx *gin.Context) {
	loginUser := models.CurrentUser(ctx, h.db)
	blogs, err := models.GetBlogList(h.db, int64(loginUser.ID))
	if err != nil {
		carrot.AbortWithJSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	for _, blog := range blogs {
		thumb, err := cache.HasThumb(blog.ID, int64(loginUser.ID), h.redis)
		if err != nil {
			carrot.AbortWithJSONError(ctx, http.StatusInternalServerError, err)
			return
		}
		blog.HasThumb = thumb
	}
	ctx.JSON(http.StatusOK, blogs)
}
