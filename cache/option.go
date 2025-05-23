package cache

import (
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init() {
	Client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

//func Get(key string) (string, error) {
//	return Client.Get(context.Background(), key).Result()
//}
//
//func Set(key string, value interface{}, expiration time.Duration) error {
//	return Client.Set(context.Background(), key, value, expiration).Err()
//}
