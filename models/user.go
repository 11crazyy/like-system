package models

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type User struct {
	ID       uint   `gorm:"primary_key" json:"id"`
	Username string `json:"username"`
}

func GetUserById(db *gorm.DB, userId uint) (*User, error) {
	var user User
	err := db.Where("id = ?", userId).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func CurrentUser(c *gin.Context, db *gorm.DB) *User {
	if cachedObj, exists := c.Get(UserField); exists && cachedObj != nil {
		return cachedObj.(*User)
	}

	session := sessions.Default(c)
	userId := session.Get(UserField)
	if userId == nil {
		return nil
	}

	user, err := GetUserById(db, userId.(uint))
	if err != nil {
		return nil
	}
	c.Set(UserField, user)
	return user
}
