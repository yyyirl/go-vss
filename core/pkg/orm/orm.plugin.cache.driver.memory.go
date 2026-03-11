package orm

import (
	"fmt"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"

	"skeyevss/core/pkg/functions"
)

var _ CacheDriver = (*CacheMemoryDriver)(nil)

type CacheMemoryDriver struct {
}

func (c *CacheMemoryDriver) Set(advanced *UseCacheAdvanced, drivers *CacheClientDriver, uniqueId string, res *CacheItem) {
	if drivers == nil || drivers.MemoryCache == nil {
		return
	}

	if advanced.Expire <= 0 {
		return
	}

	var (
		dataKey = c.cacheKeyData(advanced, uniqueId)
		sqlKey  = c.cacheKeySql(advanced, uniqueId)
		expire  = time.Duration(advanced.Expire) * time.Second
		tmp, _  = functions.ToString(res.Data)
	)
	// 设置data
	drivers.MemoryCache.Set(dataKey, res.Data, expire)
	functions.LogInfo(
		functions.SYellow("set MemoryCache, key: "),
		functions.SRed(dataKey),
		functions.SMagenta(fmt.Sprintf(" expire: %+v, data[0]: %s ...", expire, functions.TruncateString(tmp, 50))),
	)
	// 缓存有效期
	drivers.MemoryCache.Set(sqlKey, res.Sql, expire)
	functions.LogInfo(
		functions.SYellow("set MemoryCache, key: "),
		functions.SRed(sqlKey),
		functions.SMagenta(fmt.Sprintf(" expire: %s, data[1]: %+v", expire, res.Sql)),
	)

	// 记录缓存键值
	var keys []string
	if data, ok := drivers.MemoryCache.Get(advanced.CacheKeyPrefix); ok {
		keys, _ = data.([]string)
	}

	keys = functions.ArrUnique(append(keys, dataKey, sqlKey))
	drivers.MemoryCache.Set(advanced.CacheKeyPrefix, keys, cache.NoExpiration)
	functions.LogInfo(
		functions.SYellow("set MemoryCache, key: "),
		functions.SRed(advanced.CacheKeyPrefix),
		functions.SMagenta(
			fmt.Sprintf(
				" expire: %s, data[2]: %s ...",
				functions.TruncateString(strings.Join(keys, ","), 50),
				cache.NoExpiration,
			),
		),
	)
}

func (c *CacheMemoryDriver) Get(advanced *UseCacheAdvanced, drivers *CacheClientDriver, uniqueId string) []byte {
	if drivers == nil || drivers.MemoryCache == nil {
		return nil
	}

	functions.LogInfo(functions.SYellow("get MemoryCache, key: "), functions.SMagenta(c.cacheKeyData(advanced, uniqueId)))
	res, found := drivers.MemoryCache.Get(c.cacheKeyData(advanced, uniqueId))
	if found {
		b, err := functions.JSONMarshal(res)
		if err != nil {
			return nil
		}

		return b
	}

	return nil
}

func (c *CacheMemoryDriver) Delete(advanced *UseCacheAdvanced, drivers *CacheClientDriver, _ string) {
	if drivers == nil || drivers.MemoryCache == nil {
		return
	}

	if data, ok := drivers.MemoryCache.Get(advanced.CacheKeyPrefix); ok {
		if keys, ok := data.([]string); ok {
			for _, item := range keys {
				drivers.MemoryCache.Delete(item)
			}
		}
	}
}

func (c *CacheMemoryDriver) cacheKeyData(advanced *UseCacheAdvanced, uniqueId string) string {
	return advanced.CacheKeyPrefix + ":" + uniqueId + ":" + "data"
}

func (c *CacheMemoryDriver) cacheKeySql(advanced *UseCacheAdvanced, uniqueId string) string {
	return advanced.CacheKeyPrefix + ":" + uniqueId + ":" + "sql"
}

func (c *CacheMemoryDriver) currentCacheKeyPattern(advanced *UseCacheAdvanced) string {
	return advanced.CacheKeyPrefix + "*"
}
