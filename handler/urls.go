package handler

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Handlers struct {
	db    *gorm.DB
	redis *redis.Client
}
