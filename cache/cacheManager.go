package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/patrickmn/go-cache"
)

type CacheManager struct {
	hotKeyDetector *HeavyKeeper
	localCache     *cache.Cache
}

func NewCacheManager() *CacheManager {
	cm := &CacheManager{}

	// 初始化HotKey检测器
	cm.hotKeyDetector = NewHeavyKeeper(
		100,    // topK
		100000, // width
		5,      // depth
		0.92,   // decay
		10,     // minCount
	)

	cm.localCache = cache.New(
		5*time.Minute,  // 默认过期时间(对应expireAfterWrite)
		10*time.Minute, // 清理间隔
	)
	return cm
}

// GetLocalCache 获取本地缓存实例
func (cm *CacheManager) GetLocalCache() *cache.Cache {
	return cm.localCache
}

func BuildCacheKey(hashKey, key string) string {
	return hashKey + ":" + key
}

func (cm *CacheManager) Get(hashKey, key string, redis *redis.Client) any {
	compositeKey := BuildCacheKey(hashKey, key)
	localCache := cm.GetLocalCache()
	//先查本地缓存
	value, exist := localCache.Get(compositeKey)
	if exist {
		logrus.Info("本地缓存获取到数据")
		cm.hotKeyDetector.Add(key, 1) //访问次数+1
		return value
	}

	//本地缓存未命中，查询Redis
	redisValue := redis.HGet(context.Background(), hashKey, key)
	if redisValue == nil {
		return nil
	}

	addRes := cm.hotKeyDetector.Add(key, 1)
	if addRes.IsHotKey {
		//TODO 缓存到本地
		localCache.Add(compositeKey, redisValue, 0)
	}
	return redisValue
}

// 定期清理过期的热key监测数据
func (cm *CacheManager) cleanHotKeys() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.hotKeyDetector.Fading()
		}
	}
}
