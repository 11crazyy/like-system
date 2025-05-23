package handler

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"index/models"
	"net/http"
	"strconv"
)

func (h *Handlers) Login(ctx *gin.Context) {
	userId := ctx.Param("id")
	uId, _ := strconv.Atoi(userId)

	uID := uint(uId)
	user, err := models.GetUserById(h.db, uID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "用户不存在",
		})
		return
	}

	session := sessions.Default(ctx)
	session.Set(models.UserField, user.ID)
	session.Save()
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "用户存在",
		"data":    user,
	})
}
