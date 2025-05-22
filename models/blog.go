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
	CreateTime time.Time `json:"createTime" gorm:"default:CURRENT_TIMESTAMP;not null;comment:创建时间"`
	UpdateTime time.Time `json:"updateTime" gorm:"default:CURRENT_TIMESTAMP;autoUpdateTime:onUpdate;not null;comment:更新时间"`
}

// TODO 如果当前是已登陆状态，需要判断当前用户是否点赞该博客并返回给前端
func (b *Blog) GetUserThumb(db *gorm.DB, userId uint) (bool, error) {
	//var count int64
	//err := db.Model(&BlogThumb{}).Where("blog_id = ? and user_id = ?", b.ID, userId).Count(&count).Error
	//if err != nil {
	//	return false, err
	//}
	//return count > 0, nil
	return true, nil
}
