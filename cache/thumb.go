package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"index/models"
	"strconv"
)

// TODO 改成从cacheManager中获取是否点赞
func HasThumb(blogId, userId int64, redisClient *redis.Client) (bool, error) {
	//先查询布隆过滤器
	if !MayHaveThumb(userId, blogId, redisClient) {
		return false, nil //肯定没有被点赞
	}
	key := fmt.Sprintf("%s%d", models.USER_THUMB_KEY_PREFIX, userId) //一个用户对应多个博客
	exists, err := redisClient.HExists(context.Background(), key, strconv.FormatInt(blogId, 10)).Result()
	if err != nil {
		fmt.Println("redisClient.HExists error:", err)
		return false, err
	}
	return exists, nil
}

func SavaThumb(blogId, userId int64, redisClient *redis.Client) error {
	set := redisClient.HSet(context.Background(), fmt.Sprintf("%s%d", models.USER_THUMB_KEY_PREFIX, userId), strconv.FormatInt(blogId, 10), 1)
	return set.Err()
}

func DeleteThumb(blogId int64, userId int64, client *redis.Client) error {
	return client.HDel(context.Background(), fmt.Sprintf("%s%d", models.USER_THUMB_KEY_PREFIX, userId), strconv.FormatInt(blogId, 10)).Err()
}

func MayHaveThumb(userId, blogId int64, redisClient *redis.Client) bool {
	key := fmt.Sprintf("%s%d", models.USER_THUMB_BLOOM_KEY, userId)
	exists, err := redisClient.BFExists(context.Background(), key, strconv.FormatInt(blogId, 10)).Result()
	if err != nil {
		fmt.Println("BFExists error:", err)
		return false
	}
	return exists
}
