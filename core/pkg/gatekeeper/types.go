// @Title        types
// @Description  main
// @Create       yiyiyi 2025/9/26 13:35

package gatekeeper

import (
	"skeyevss/core/pkg/redis"

	"sync"
)

type (
	Gatekeeper struct {
		RedisClient *redis.GoRedisClient
		Key         string
		Node        string
		Expire      uint64
		stopChan    chan struct{}
		once        sync.Once
		rateLimit   int64 // 每秒允许的请求数
	}

	CacheItem struct {
		ID     string `json:"id"`
		Expire uint64 `json:"expire"`
	}
)

const (
	cacheKeyPrefix = "gatekeeper"

	prefix = "=GB=l/yb=Gy/b"
	suffix = "=/b/===/=b"

	idCacheKey        = cacheKeyPrefix + ":identity" // 记录id对应的token
	uniqueIdsCacheKey = cacheKeyPrefix + ":uniqueId" // 生成id

	produceIDRedisLockKey = "produce-id-redis-lock"
)
