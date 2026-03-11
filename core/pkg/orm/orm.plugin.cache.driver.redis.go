package orm

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis"

	"skeyevss/core/pkg/functions"
)

var _ CacheDriver = (*CacheRedisDriver)(nil)

type CacheRedisDriver struct {
}

func (c *CacheRedisDriver) Set(advanced *UseCacheAdvanced, drivers *CacheClientDriver, uniqueId string, res *CacheItem) {
	if drivers == nil || drivers.RedisClient == nil {
		return
	}

	if advanced.Expire <= 0 {
		return
	}

	var data = res.Data
	if ok, _ := functions.IsSimpleType(res.Data); !ok {
		b, _ := json.Marshal(res.Data)
		data = string(b)
	}

	if _, err := drivers.RedisClient.Set(c.cacheKeyData(advanced, uniqueId), data, time.Duration(advanced.Expire)*time.Second); err != nil {
		functions.LogError(err)
		return
	}

	if _, err := drivers.RedisClient.Set(c.cacheKeySql(advanced, uniqueId), res.Sql, time.Duration(advanced.Expire)*time.Second); err != nil {
		functions.LogError(err)
		return
	}
}

func (c *CacheRedisDriver) Get(advanced *UseCacheAdvanced, drivers *CacheClientDriver, uniqueId string) []byte {
	if drivers == nil || drivers.RedisClient == nil {
		return nil
	}

	res, err := drivers.RedisClient.Get(c.cacheKeyData(advanced, uniqueId))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}

		functions.LogError(err)
		return nil
	}

	return res
}

func (c *CacheRedisDriver) Delete(advanced *UseCacheAdvanced, drivers *CacheClientDriver, _ string) {
	if drivers == nil || drivers.RedisClient == nil {
		return
	}

	keys, err := drivers.RedisClient.Scan(c.currentCacheKeyPattern(advanced))
	if err != nil {
		functions.LogError(err)
		return
	}

	if len(keys) <= 0 {
		return
	}

	if _, err := drivers.RedisClient.Del(keys...); err != nil {
		functions.LogError(err)
	}
}

func (c *CacheRedisDriver) cacheKeyData(advanced *UseCacheAdvanced, uniqueId string) string {
	return advanced.CacheKeyPrefix + ":" + uniqueId + ":" + "data"
}

func (c *CacheRedisDriver) cacheKeySql(advanced *UseCacheAdvanced, uniqueId string) string {
	return advanced.CacheKeyPrefix + ":" + uniqueId + ":" + "sql"
}

func (c *CacheRedisDriver) currentCacheKeyPattern(advanced *UseCacheAdvanced) string {
	return advanced.CacheKeyPrefix + ":*"
}
