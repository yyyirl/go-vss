/**
 * @Author:         yi
 * @Description:    foundations
 * @Version:        1.0.0
 * @Date:           2022/10/13 11:24
 */
package redis

import (
	r "github.com/go-redis/redis"
	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/pkg/redis"
	"skeyevss/core/tps"
)

const cachePrefix = "sk-"

var RedisNil = r.Nil

type Client struct {
	// github.com/gomodule/redigo/redis
	// *redis.RedisGoClient

	// github.com/go-redis/redis
	*redis.GoRedisClient
}

func New(mode string, logEncoding string, redisConf tps.YamlRedis, log logx.LogConf) *Client {
	return &Client{
		// RedisGoClient: redis.NewRedisGoClient(mode, redisConf, log),
		GoRedisClient: redis.NewGoRedisClient(mode, logEncoding, redisConf, log),
	}
}
