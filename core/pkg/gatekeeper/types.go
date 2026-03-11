// @Title        types
// @Description  main
// @Create       yiyiyi 2025/9/26 13:35

package gatekeeper

import "skeyevss/core/pkg/redis"

type (
	CacheItem struct {
		ID     string
		Expire uint64
	}

	Gatekeeper struct {
		RedisClient *redis.GoRedisClient
		Key         string // 秘钥
		Node        string
		Expire      uint64 // 过期时间 毫秒
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
