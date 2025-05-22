package models

import "time"

type Thumb struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	UserID     int64     `json:"userId" gorm:"not null"`
	BlogID     int64     `json:"blogId" gorm:"not null"`
	CreateTime time.Time `json:"createTime" gorm:"default:CURRENT_TIMESTAMP;not null;comment:创建时间"`
}
