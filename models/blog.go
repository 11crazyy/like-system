package models

import (
	"gorm.io/gorm"
	"time"
)

type Blog struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	UserID     int64     `json:"userId" gorm:"not null"`
	Title      *string   `json:"title,omitempty" gorm:"size:512;comment:标题"`
	CoverImg   *string   `json:"coverImg,omitempty" gorm:"size:1024;comment:封面"`
	Content    string    `json:"content" gorm:"not null;comment:内容"`
	ThumbCount int       `json:"thumbCount" gorm:"default:0;not null;comment:点赞数"`
	HasThumb   bool      `json:"hasThumb"`
	CreateTime time.Time `json:"createTime" gorm:"default:CURRENT_TIMESTAMP;not null;comment:创建时间"`
	UpdateTime time.Time `json:"updateTime" gorm:"default:CURRENT_TIMESTAMP;autoUpdateTime:onUpdate;not null;comment:更新时间"`
}

func GetBlogById(db *gorm.DB, blogId int64) (*Blog, error) {
	var blog Blog
	err := db.Where("id = ?", blogId).First(&blog).Error
	return &blog, err
}

func UpdateThumbNum(db *gorm.DB, blogId string, thumbCount int) error {
	if err := db.Model(&Blog{}).Where("id = ?", blogId).Update("thumb_count", thumbCount).Error; err != nil {
		return err
	}
	return nil
}

func GetBlogList(db *gorm.DB, userId int64) ([]*Blog, error) {
	var blogs []*Blog
	err := db.Where("user_id = ?", userId).Find(&blogs).Error
	if err != nil {
		return nil, err
	}
	return blogs, nil
}
