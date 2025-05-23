package models

import (
	"gorm.io/gorm"
	"time"
)

type Thumb struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	UserID     int64     `json:"userId" gorm:"not null"`
	BlogID     int64     `json:"blogId" gorm:"not null"`
	CreateTime time.Time `json:"createTime" gorm:"default:CURRENT_TIMESTAMP;not null;comment:创建时间"`
}

func GetBlogByUserIdAndBlogId(db *gorm.DB, blogId string, userId string) (error, int64) {
	var count int64
	err := db.Model(&Thumb{}).Where("user_id = ? AND blog_id = ?", userId, blogId).Count(&count).Error
	if err != nil {
		return err, 0
	}
	return nil, count
}

func CreateThumb(db *gorm.DB, thumb *Thumb) error {
	return db.Create(thumb).Error
}

func DeleteThumb(db *gorm.DB, blogId, userId string) error {
	if err := db.Where("user_id = ? AND blog_id = ?", userId, blogId).Delete(&Thumb{}).Error; err != nil {
		return err
	}
	return nil
}

// 是否已点赞
func HasThumb(db *gorm.DB, blogId, userId int64) {

}
