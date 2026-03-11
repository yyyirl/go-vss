/**
 * @Author:         yi
 * @Description:    github.com/go-redis/redis
 * @Version:        1.0.0
 * @Date:           2023/6/30 11:58
 */
package redis

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-redis/redis"
	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/tps"
)

const redisExecuteLogPrefix = "go-redis execute commend: "

type GoRedisClient struct {
	// 是否是集群
	isCluster bool
	// redis客户端 单机
	client *redis.Client
	// 集群
	clusterClient *redis.ClusterClient

	logEncoding string
}

func NewGoRedisClient(_ string, logEncoding string, redisConf tps.YamlRedis, _ logx.LogConf) *GoRedisClient {
	var isCluster = len(redisConf.Hosts) > 0
	if isCluster {
		var client = redis.NewClusterClient(
			&redis.ClusterOptions{
				Addrs:        redisConf.Hosts,
				Password:     redisConf.Pass,
				DialTimeout:  200 * time.Microsecond, // 设置连接超时
				ReadTimeout:  200 * time.Microsecond, // 设置读取超时
				WriteTimeout: 200 * time.Microsecond, // 设置写入超时
			},
		)

		if _, err := client.Ping().Result(); err != nil {
			panic(fmt.Errorf("redis 链接失败 %s", err))
		}

		return &GoRedisClient{
			isCluster: isCluster,
			// 集群
			clusterClient: client,
			logEncoding:   logEncoding,
		}
	}

	var client = redis.NewClient(
		&redis.Options{
			Addr:     redisConf.Host,
			Password: redisConf.Pass,
		},
	)

	if _, err := client.Ping().Result(); err != nil {
		panic(fmt.Errorf("redis 链接失败 %s", err))
	}

	return &GoRedisClient{
		isCluster: isCluster,
		// 单机
		client:      client,
		logEncoding: logEncoding,
	}
}

func (r *GoRedisClient) log(cmd string) {
	if r.logEncoding == "json" {
		logx.Info("response: " + cmdInfo(cmd))
		return
	}

	logx.Info(functions.SYellow(redisExecuteLogPrefix), "response: "+functions.SMagenta(cmdInfo(cmd)))
}

// 设置key value
func (r *GoRedisClient) Set(key string, data interface{}, expire time.Duration) (string, error) {
	var resp *redis.StatusCmd
	if r.isCluster {
		resp = r.clusterClient.Set(key, data, expire)
	} else {
		resp = r.client.Set(key, data, expire)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return "", err
	}

	return b, err
}

// 获取缓存
func (r *GoRedisClient) Get(key string) ([]byte, error) {
	var resp *redis.StringCmd
	if r.isCluster {
		resp = r.clusterClient.Get(key)
	} else {
		resp = r.client.Get(key)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return nil, err
	}

	return []byte(b), err
}

// 获取缓存 keys
// WARNING:
// - keys 命令会一次性返回所有匹配的键，这可能导致 Redis 阻塞，严重影响线上服务的稳定性
// - 请使用Scan方法
func (r *GoRedisClient) Keys(key string) ([]string, error) {
	var resp *redis.StringSliceCmd
	if r.isCluster {
		resp = r.clusterClient.Keys(key)
	} else {
		resp = r.client.Keys(key)
	}

	r.log(resp.String())
	return resp.Result()
}

// 获取缓存 keys
func (r *GoRedisClient) Scan(key string) ([]string, error) {
	var (
		cursor uint64
		keys   []string
	)
	for {
		var (
			resp  *redis.ScanCmd
			_keys []string
			err   error
		)
		if r.isCluster {
			resp = r.clusterClient.Scan(cursor, key, 0)
		} else {
			resp = r.client.Scan(cursor, key, 0)
		}

		r.log(resp.String())
		_keys, cursor, err = resp.Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, _keys...)
		if cursor <= 0 {
			break
		}
	}

	return keys, nil
}

// 获取多个
func (r *GoRedisClient) MGet(keys []string) ([]interface{}, error) {
	var resp *redis.SliceCmd
	if r.isCluster {
		resp = r.clusterClient.MGet(keys...)
	} else {
		resp = r.client.MGet(keys...)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return nil, err
	}

	return b, err
}

// 获取int64
func (r *GoRedisClient) GetInt64(key string) (int64, error) {
	var resp *redis.StringCmd
	if r.isCluster {
		resp = r.clusterClient.Get(key)
	} else {
		resp = r.client.Get(key)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return 0, err
	}

	num, err := strconv.ParseInt(b, 10, 64)
	if err != nil {
		return 0, err
	}

	return num, err
}

// 将一个或多个值插入到列表的尾部(最右边)
func (r *GoRedisClient) RPush(key string, data interface{}) (int, error) {
	value, err := functions.JSONMarshal(data)
	if err != nil {
		return 0, err
	}

	var resp *redis.IntCmd
	if r.isCluster {
		resp = r.clusterClient.RPush(key, value)
	} else {
		resp = r.client.RPush(key, value)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return 0, err
	}

	return int(b), err
}

// 将一个或多个值插入到列表的尾部(最右边)
func (r *GoRedisClient) RSimplePush(key string, data interface{}) (int, error) {
	var resp *redis.IntCmd
	if r.isCluster {
		resp = r.clusterClient.RPush(key, data)
	} else {
		resp = r.client.RPush(key, data)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return 0, err
	}

	return int(b), err
}

// 将一个或多个值插入到列表的尾部(最右边)
func (r *GoRedisClient) RSlicePush(key string, data []interface{}) (int, error) {
	var resp *redis.IntCmd
	if r.isCluster {
		resp = r.clusterClient.RPush(key, data...)
	} else {
		resp = r.client.RPush(key, data...)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return 0, err
	}

	return int(b), err
}

// 将一个或多个成员元素加入到集合中，已经存在于集合的成员元素将被忽略
func (r *GoRedisClient) SAdd(key string, data ...interface{}) (int, error) {
	var resp *redis.IntCmd
	if r.isCluster {
		resp = r.clusterClient.SAdd(key, data...)
	} else {
		resp = r.client.SAdd(key, data...)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return 0, err
	}

	return int(b), err
}

// 获取集合成员数量
func (r *GoRedisClient) SCard(key string) (int64, error) {
	var resp *redis.IntCmd
	if r.isCluster {
		resp = r.clusterClient.SCard(key)
	} else {
		resp = r.client.SCard(key)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return 0, err
	}

	return b, err
}

// 将一个或多个成员元素加入到集合中，已经存在于集合的成员元素将被忽略
func (r *GoRedisClient) SMAdd(key string, data ...interface{}) (int, error) {
	var resp *redis.IntCmd
	if r.isCluster {
		resp = r.clusterClient.SAdd(key, data...)
	} else {
		resp = r.client.SAdd(key, data...)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return 0, err
	}

	return int(b), err
}

// 移除集合中的一个或多个成员元素，不存在的成员元素会被忽略
func (r *GoRedisClient) SRem(key string, data ...interface{}) (int, error) {
	var resp *redis.IntCmd
	if r.isCluster {
		resp = r.clusterClient.SRem(key, data...)
	} else {
		resp = r.client.SRem(key, data...)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return 0, err
	}

	return int(b), err
}

// 判断成员元素是否是集合的成员
func (r *GoRedisClient) SIsMember(key string, data interface{}) (bool, error) {
	var resp *redis.BoolCmd
	if r.isCluster {
		resp = r.clusterClient.SIsMember(key, data)
	} else {
		resp = r.client.SIsMember(key, data)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return false, err
	}

	return b, err
}

// 检查键值是否存在
func (r *GoRedisClient) Exists(key string) bool {
	var resp *redis.IntCmd
	if r.isCluster {
		resp = r.clusterClient.Exists(key)
	} else {
		resp = r.client.Exists(key)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return false
	}

	return b > 0
}

// 获取缓存生存时间
func (r *GoRedisClient) GetTTL(key string) (time.Duration, error) {
	var resp *redis.DurationCmd
	if r.isCluster {
		resp = r.clusterClient.TTL(key)
	} else {
		resp = r.client.TTL(key)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return 0, err
	}

	return b, err
}

// 删除键
func (r *GoRedisClient) Del(key ...string) (int64, error) {
	var resp *redis.IntCmd
	if r.isCluster {
		resp = r.clusterClient.Del(key...)
	} else {
		resp = r.client.Del(key...)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return 0, err
	}

	return b, err
}

// 设置 hash redis
func (r *GoRedisClient) HSet(marsh bool, key, field string, data interface{}) (bool, error) {
	if marsh {
		value, err := functions.JSONMarshal(data)
		if err != nil {
			return false, err
		}
		data = value
	}

	var resp *redis.BoolCmd
	if r.isCluster {
		resp = r.clusterClient.HSet(key, field, data)
	} else {
		resp = r.client.HSet(key, field, data)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return false, err
	}

	return b, err
}

func (r *GoRedisClient) HMSet(key string, fields map[string]interface{}) (string, error) {
	var resp *redis.StatusCmd
	if r.isCluster {
		resp = r.clusterClient.HMSet(key, fields)
	} else {
		resp = r.client.HMSet(key, fields)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return "", err
	}

	return b, err
}

func (r *GoRedisClient) HIncrBy(key, field string, num int64) (int64, error) {
	var resp *redis.IntCmd
	if r.isCluster {
		resp = r.clusterClient.HIncrBy(key, field, num)
	} else {
		resp = r.client.HIncrBy(key, field, num)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return 0, err
	}

	return b, err
}

func (r *GoRedisClient) HGetStr(key, field string) (string, error) {
	var resp *redis.StringCmd
	if r.isCluster {
		resp = r.clusterClient.HGet(key, field)
	} else {
		resp = r.client.HGet(key, field)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return "", err
	}

	return b, err
}

func (r *GoRedisClient) HKeys(key string) ([]string, error) {
	var resp *redis.StringSliceCmd
	if r.isCluster {
		resp = r.clusterClient.HKeys(key)
	} else {
		resp = r.client.HKeys(key)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return nil, err
	}

	return b, err
}

func (r *GoRedisClient) HGet(key, field string) ([]byte, error) {
	res, err := r.HGetStr(key, field)
	if err != nil {
		return nil, err
	}

	return []byte(res), err
}

func (r *GoRedisClient) HGetInt64(key, field string) (int64, error) {
	res, err := r.HGetStr(key, field)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(res, 10, 64)
}

func (r *GoRedisClient) HGetBool(key, field string) (bool, error) {
	var resp *redis.BoolCmd
	if r.isCluster {
		resp = r.clusterClient.HExists(key, field)
	} else {
		resp = r.client.HExists(key, field)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return false, err
	}

	return b, err
}

func (r *GoRedisClient) HGetFloat64(key, field string) (float64, error) {
	res, err := r.HGetStr(key, field)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(res, 64)
}

func (r *GoRedisClient) HGetAll(key string) (map[string]string, error) {
	var (
		cursor uint64
		result = make(map[string]string)
	)
	for {
		var resp *redis.ScanCmd
		if r.isCluster {
			resp = r.clusterClient.HScan(key, cursor, "*", 100)
		} else {
			resp = r.client.HScan(key, cursor, "*", 100)
		}

		fields, nextCursor, err := resp.Result()
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				result[fields[i]] = fields[i+1]
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return result, nil
}

func (r *GoRedisClient) HMGet(key string, fields ...string) ([]interface{}, error) {
	var resp *redis.SliceCmd
	if r.isCluster {
		resp = r.clusterClient.HMGet(key, fields...)
	} else {
		resp = r.client.HMGet(key, fields...)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return nil, err
	}

	return b, err
}

func (r *GoRedisClient) HDel(key string, fields ...string) (int64, error) {
	var resp *redis.IntCmd
	if r.isCluster {
		resp = r.clusterClient.HDel(key, fields...)
	} else {
		resp = r.client.HDel(key, fields...)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return 0, err
	}

	return b, err
}

func (r *GoRedisClient) LTrim(key string, start, end int) (string, error) {
	var resp *redis.StatusCmd
	if r.isCluster {
		resp = r.clusterClient.LTrim(key, int64(start), int64(end))
	} else {
		resp = r.client.LTrim(key, int64(start), int64(end))
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return "", err
	}

	return b, err
}

// 获取列表
func (r *GoRedisClient) LRange(key string, start int, end int) ([]string, error) {
	var resp *redis.StringSliceCmd
	if r.isCluster {
		resp = r.clusterClient.LRange(key, int64(start), int64(end))
	} else {
		resp = r.client.LRange(key, int64(start), int64(end))
	}

	// r.log(String())
	b, err := resp.Result()
	if err != nil {
		return nil, err
	}

	return b, err
}

func CheckNil(err error) bool {
	return err == redis.Nil
}

func cmdInfo(content string) string {
	if utf8.RuneCountInString(content) <= 100 {
		return content
	}

	return string([]rune(content)[:100]) + " ..."
}

func (r *GoRedisClient) SetNX(key string, data interface{}, expire time.Duration) (bool, error) {
	var resp *redis.BoolCmd
	if r.isCluster {
		resp = r.clusterClient.SetNX(key, data, expire)
	} else {
		resp = r.client.SetNX(key, data, expire)
	}

	r.log(resp.String())
	b, err := resp.Result()
	if err != nil {
		return false, err
	}

	return b, err
}

func (r *GoRedisClient) evalCommand(script string, keys []string, args ...interface{}) string {
	var (
		formattedScript = strings.Join(strings.Fields(strings.ReplaceAll(strings.TrimSpace(script), "\n", " ")), " ")
		keysStr         = fmt.Sprintf("%d", len(keys))
		argsStr         = ""
	)
	for _, key := range keys {
		keysStr += " " + key
	}

	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			argsStr += " " + v
		case int, int64:
			argsStr += " " + fmt.Sprintf("%d", v)
		default:
			argsStr += " " + fmt.Sprintf("%v", v)
		}
	}

	return fmt.Sprintf("EVAL \"%s\" %s%s", formattedScript, keysStr, argsStr)
}

func (r *GoRedisClient) Eval(script string, keys []string, args ...interface{}) (interface{}, error) {
	var resp *redis.Cmd
	if r.isCluster {
		resp = r.clusterClient.Eval(script, keys, args...)
	} else {
		resp = r.client.Eval(script, keys, args...)
	}

	r.log(r.evalCommand(script, keys, args...))
	res, err := resp.Result()
	if err != nil {
		return nil, err
	}

	return res, err
}

// 获取锁
func (r *GoRedisClient) AcquireLock(lockKey string, node string, ttl time.Duration) (bool, error) {
	result, err := r.Eval(
		`if redis.call('SET', KEYS[1], ARGV[1], 'NX', 'PX', ARGV[2]) then
	return 1
else
	return 0
end`,
		[]string{lockKey},
		node,
		ttl.Milliseconds(),
	)
	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}

// 释放锁
func (r *GoRedisClient) ReleaseLock(lockKey string, node string) (bool, error) {
	result, err := r.Eval(
		`if redis.call('GET', KEYS[1]) == ARGV[1] then
	return redis.call('DEL', KEYS[1])
else
	return 0
end`,
		[]string{lockKey},
		node,
	)
	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}
