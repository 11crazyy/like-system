package models

import "gorm.io/gorm"

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
