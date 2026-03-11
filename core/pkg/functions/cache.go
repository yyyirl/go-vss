/**
 * @Author:         yi
 * @Description:    cache
 * @Version:        1.0.0
 * @Date:           2022/10/10 12:15
 */
package functions

import (
	"time"

	"github.com/allegro/bigcache"
	cmap "github.com/orcaman/concurrent-map"

	"skeyevss/core/constants"
)

var (
	simpleCachePool       cmap.ConcurrentMap
	simpleExpireCachePool *bigcache.BigCache
)

func init() {
	simpleCachePool = cmap.New()
}

func CachePoolCall(prefix string, s string, call func() interface{}) interface{} {
	var key = prefix + "-" + s
	val, ok := simpleCachePool.Get(key)
	if ok {
		return val
	}

	var data = call()
	if val == nil {
		simpleCachePool.Set(key, data)
	}

	return data
}

func CachePoolWithExpireCall(key string, expire int64, val interface{}, call func() (interface{}, error)) error {
	var err error
	if simpleExpireCachePool == nil {
		// 过期时间为 expire * 2
		simpleExpireCachePool, err = BigCacheThrottle(time.Duration(expire) * time.Second)
		if err != nil {
			return err
		}
	}

	data, err := simpleExpireCachePool.Get(key)
	if err != nil && err.Error() != constants.ERR_LOCAL_CACHE_NOTFOUND {
		return err
	}

	if data == nil {
		resp, err := call()
		if err != nil {
			return err
		}

		data, err = JSONMarshal(resp)
		if err != nil {
			return err
		}

		err = simpleExpireCachePool.Set(key, data)
		if err != nil {
			return err
		}
	}

	if err := JSONUnmarshal(data, val); err != nil {
		return nil
	}

	return nil
}

func BigCacheThrottle(alive time.Duration) (*bigcache.BigCache, error) {
	return bigcache.NewBigCache(
		bigcache.Config{
			// 设置分区的数量，必须是2的整倍数
			Shards: 1024,
			// LifeWindow后,缓存对象被认为不活跃,但并不会删除对象
			LifeWindow: alive,
			// CleanWindow后，会删除被认为不活跃的对象，0代表不操作；
			CleanWindow: 10 * time.Second,
			// 设置最大存储对象数量，仅在初始化时可以设置
			// MaxEntriesInWindow: 1000 * 10 * 60,
			MaxEntriesInWindow: 1,
			// 缓存对象的最大字节数，仅在初始化时可以设置
			MaxEntrySize: 500,
			// 是否打印内存分配信息
			Verbose: true,
			// 设置缓存最大值(单位为MB),0表示无限制
			HardMaxCacheSize: 8192,
			// 在缓存过期或者被删除时,可设置回调函数，参数是(key、val)，默认是nil不设置
			// OnRemove: func(key string, entry []byte) {},
			// 在缓存过期或者被删除时,可设置回调函数，参数是(key、val,reason)默认是nil不设置
			OnRemoveWithReason: nil,
		},
	)
}

func BigCachePermanent() (*bigcache.BigCache, error) {
	config := bigcache.Config{
		Shards:             1024,
		LifeWindow:         60 * time.Minute,
		CleanWindow:        0,
		MaxEntriesInWindow: 1,
		MaxEntrySize:       500,
		Verbose:            true,
		HardMaxCacheSize:   0,
		OnRemoveWithReason: nil,
	}

	return bigcache.NewBigCache(config)
}
