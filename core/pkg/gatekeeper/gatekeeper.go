// @Title        gatekeeper
// @Description  守卫，用于生成和管理访问凭证，提供黑名单和限流功能
// @Create       yiyiyi 2025/9/26 13:44

package gatekeeper

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/redis"
)

const (
	defaultCleanupInterval = 10 * time.Second
	maxRetryCount          = 100
	blacklistKey           = "gatekeeper:blacklist"
	rateLimitKey           = "gatekeeper:ratelimit"
	defaultRateLimit       = 100 // 默认每秒请求数
)

var (
	// 确保清理过期token的goroutine只启动一次
	clearExpireToken sync.Once
	// 清理操作的锁通道，防止并发清理导致的问题
	clearLockChan = make(chan struct{}, 1)
	// 并发控制通道，用于处理ID生成时的并发冲突
	concurrenceLockChan = make(chan struct{}, 1)
)

// BlacklistItem 黑名单条目结构
type BlacklistItem struct {
	ID        string `json:"id"`         // 被加入黑名单的ID
	Reason    string `json:"reason"`     // 加入原因
	Expire    uint64 `json:"expire"`     // 过期时间（毫秒时间戳）
	CreatedAt uint64 `json:"created_at"` // 创建时间（毫秒时间戳）
}

// New 创建Gatekeeper实例
// redisClient: Redis客户端实例
// expire: 凭证有效期（毫秒）
// key: 加密密钥
// node: 节点标识，用于生成唯一ID
func New(redisClient *redis.GoRedisClient, expire uint64, key, node string) *Gatekeeper {
	if redisClient == nil {
		panic("redis client cannot be nil")
	}

	if key == "" {
		panic("encryption key cannot be empty")
	}

	var instance = &Gatekeeper{
		RedisClient: redisClient,
		Key:         key,
		Node:        node,
		Expire:      expire,
		stopChan:    make(chan struct{}),
		rateLimit:   defaultRateLimit,
	}
	// 启动后台清理协程
	clearExpireToken.Do(func() {
		go instance.startCleanupRoutine()
	})

	return instance
}

// SetRateLimit 设置限流阈值
// limit: 每秒允许的请求数
func (l *Gatekeeper) SetRateLimit(limit int64) {
	if limit > 0 {
		l.rateLimit = limit
	}
}

// Stop 停止后台清理协程
func (l *Gatekeeper) Stop() {
	l.once.Do(func() { close(l.stopChan) })
}

// startCleanupRoutine 启动后台清理协程
func (l *Gatekeeper) startCleanupRoutine() {
	var ticker = time.NewTicker(defaultCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-l.stopChan:
			return

		case t := <-ticker.C:
			l.doClearExpireToken(t)
			l.cleanupExpiredBlacklist()
		}
	}
}

// doClearExpireToken 过期token清理
func (l *Gatekeeper) doClearExpireToken(t time.Time) {
	// 使用select避免阻塞
	select {
	case clearLockChan <- struct{}{}:
		defer func() { <-clearLockChan }()

	default:
		// 如果无法获取锁，直接返回避免阻塞
		return
	}

	maps, err := l.RedisClient.HGetAll(idCacheKey)
	if err != nil {
		functions.LogError("fetch all expire token error:", err)
		return
	}

	if len(maps) == 0 {
		return
	}

	var (
		deleteIds   = make([]string, 0)
		currentTime = uint64(t.UnixMilli())
	)
	for key, item := range maps {
		data, err := l.decrypt(item)
		if err != nil {
			// 解密失败，删除该条目
			deleteIds = append(deleteIds, key)
			continue
		}

		if currentTime > data.Expire {
			deleteIds = append(deleteIds, key)
		}
	}

	if len(deleteIds) == 0 {
		return
	}

	deleteIds = functions.ArrUnique(deleteIds)
	// 删除过期token
	if _, err = l.RedisClient.HDel(idCacheKey, deleteIds...); err != nil {
		functions.LogError("delete expire token error:", err)
	}

	// 从唯一ID集合中删除
	if _, err = l.RedisClient.SRem(uniqueIdsCacheKey, functions.SliceToSliceAny(deleteIds)...); err != nil {
		functions.LogError("delete expire id error:", err)
	}
}

// AddToBlacklist 将ID加入黑名单
// id: 要加入黑名单的ID
// reason: 加入原因
// duration: 黑名单持续时间
func (l *Gatekeeper) AddToBlacklist(id, reason string, duration time.Duration) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	data, err := functions.JSONMarshal(&BlacklistItem{
		ID:        id,
		Reason:    reason,
		Expire:    uint64(time.Now().Add(duration).UnixMilli()),
		CreatedAt: uint64(time.Now().UnixMilli()),
	})
	if err != nil {
		return fmt.Errorf("marshal blacklist item error: %s", err)
	}

	var key = fmt.Sprintf("%s:%s", blacklistKey, id)
	if _, err := l.RedisClient.Set(key, string(data), duration); err != nil {
		return fmt.Errorf("set blacklist error: %s", err)
	}

	return nil
}

// RemoveFromBlacklist 从黑名单中移除
// id: 要从黑名单移除的ID
func (l *Gatekeeper) RemoveFromBlacklist(id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	_, err := l.RedisClient.Del(fmt.Sprintf("%s:%s", blacklistKey, id))
	return err
}

// IsBlacklisted 检查是否在黑名单中
// id: 要检查的ID
// 返回值: 是否在黑名单中, 黑名单条目信息, 错误信息
func (l *Gatekeeper) IsBlacklisted(id string) (bool, *BlacklistItem, error) {
	if id == "" {
		return false, nil, errors.New("id cannot be empty")
	}

	data, err := l.RedisClient.Get(fmt.Sprintf("%s:%s", blacklistKey, id))
	if err != nil {
		if redis.CheckNil(err) {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("get blacklist error: %s", err)
	}

	var item BlacklistItem
	if err = functions.JSONUnmarshal(data, &item); err != nil {
		return false, nil, fmt.Errorf("unmarshal blacklist item error: %s", err)
	}

	var now = uint64(time.Now().UnixMilli())
	if now > item.Expire {
		// 异步清理过期黑名单
		go func() {
			if err := l.RemoveFromBlacklist(id); err != nil {
				functions.LogError("remove blacklist error:", err)
			}
		}()

		return false, nil, nil
	}

	return true, &item, nil
}

// cleanupExpiredBlacklist 清理过期黑名单
func (l *Gatekeeper) cleanupExpiredBlacklist() {
	keys, err := l.RedisClient.Keys(blacklistKey + ":*")
	if err != nil {
		functions.LogError("fetch blacklist keys error:", err)
		return
	}

	var (
		now        = uint64(time.Now().UnixMilli())
		deleteKeys = make([]string, 0)
	)
	for _, key := range keys {
		data, err := l.RedisClient.Get(key)
		if err != nil {
			continue
		}

		var item BlacklistItem
		if err = functions.JSONUnmarshal(data, &item); err != nil {
			deleteKeys = append(deleteKeys, key)
			continue
		}

		if now > item.Expire {
			deleteKeys = append(deleteKeys, key)
		}
	}

	if len(deleteKeys) > 0 {
		_, _ = l.RedisClient.Del(deleteKeys...)
	}
}

// RateLimit 限流检查
// id: 要进行限流检查的ID
// 返回值: 是否允许通过, 当前请求数, 错误信息
func (l *Gatekeeper) RateLimit(id string) (bool, int64, error) {
	if id == "" {
		return false, 0, errors.New("id cannot be empty")
	}

	var (
		now = time.Now()
		// 使用Lua脚本实现原子操作
		script = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local now = tonumber(ARGV[2])

local current = redis.call('INCR', key)
if current == 1 then
	redis.call('EXPIRE', key, 1)
end

if current > limit then
	return {0, current}
end

return {1, current}
`
	)

	result, err := l.RedisClient.Eval(
		script,
		[]string{
			fmt.Sprintf(
				"%s:%d",
				fmt.Sprintf("%s:%s", rateLimitKey, id),
				now.Truncate(time.Second).Unix(),
			),
		},
		l.rateLimit,
		now.Unix(),
	)
	if err != nil {
		return false, 0, fmt.Errorf("rate limit error: %s", err)
	}

	v, ok := result.([]interface{})
	if !ok && len(v) != 2 {
		return false, 0, errors.New("invalid rate limit result")
	}

	allowed, err := functions.InterfaceToNumber[int64](v[0])
	if err != nil {
		return false, 0, errors.New("invalid rate limit result [0]")
	}

	current, err := functions.InterfaceToNumber[int64](v[1])
	if err != nil {
		return false, 0, errors.New("invalid rate limit result [1]")
	}

	return allowed == 1, current, nil
}

// GetRateLimitStatus 获取限流状态
// id: 要查询的ID
// 返回值: 当前请求数, 限制数, 重置时间, 错误信息
func (l *Gatekeeper) GetRateLimitStatus(id string) (current int64, limit int64, resetTime time.Time, err error) {
	if id == "" {
		return 0, 0, time.Time{}, errors.New("id cannot be empty")
	}

	var (
		key    = fmt.Sprintf("%s:%s", rateLimitKey, id)
		window = time.Now().Truncate(time.Second)
	)
	value, err := l.RedisClient.Get(fmt.Sprintf("%s:%d", key, window.Unix()))
	if err != nil {
		if redis.CheckNil(err) {
			return 0, l.rateLimit, window.Add(time.Second), nil
		}

		return 0, 0, time.Time{}, err
	}

	currentVal, _ := strconv.ParseInt(string(value), 10, 64)
	return currentVal, l.rateLimit, window.Add(time.Second), nil
}

// ClientID 生成唯一的客户端ID
func (l *Gatekeeper) ClientID() (string, error) {
	return l.clientID(0, "")
}

// clientID 递归生成唯一ID，处理冲突
func (l *Gatekeeper) clientID(counter int, desc string) (string, error) {
	if counter >= maxRetryCount {
		return "", fmt.Errorf("retry counter out of range[%s]", desc)
	}

	// 生成唯一ID
	var id = l.generateUniqueID()

	// 检查ID是否已存在
	exists, err := l.RedisClient.SIsMember(uniqueIdsCacheKey, id)
	if err != nil {
		return "", fmt.Errorf("check id existence error: %s", err)
	}

	if exists {
		// 使用通道进行简单的并发控制
		select {
		case concurrenceLockChan <- struct{}{}:
			time.Sleep(10 * time.Millisecond)
			<-concurrenceLockChan

		default:
		}

		return l.clientID(counter+1, "already exists. retry")
	}

	// 将ID添加到集合中
	if _, err = l.RedisClient.SAdd(uniqueIdsCacheKey, id); err != nil {
		// 继续执行，因为添加失败可能只是缓存问题
		functions.LogError("create unique cache id error:", err)
	}

	return id, nil
}

// generateUniqueID 生成唯一ID
// 组合时间戳和节点信息，降低冲突概率
func (l *Gatekeeper) generateUniqueID() string {
	// 组合时间戳和节点信息，降低冲突概率
	var nodeHash = fmt.Sprintf("%x", l.Node)
	if len(nodeHash) > 8 {
		nodeHash = nodeHash[:8]
	}
	return fmt.Sprintf(
		"%s_%s_%d",
		strconv.FormatInt(time.Now().UnixNano(), 36),
		nodeHash,
		time.Now().UnixNano(),
	)
}

// Credential 生成凭证
// id: 客户端ID
// 返回值: 生成的凭证token, 错误信息
func (l *Gatekeeper) Credential(id string) (string, error) {
	if id == "" {
		return "", errors.New("id cannot be empty")
	}

	// 获取锁
	select {
	case clearLockChan <- struct{}{}:
		defer func() { <-clearLockChan }()

	default:
		return "", errors.New("system busy, try again later")
	}

	// 创建缓存项
	token, err := l.encryption(&CacheItem{
		ID:     id,
		Expire: uint64(time.Now().UnixMilli()) + l.Expire,
	})
	if err != nil {
		return "", fmt.Errorf("encryption error: %s", err)
	}

	// 存储凭证
	if _, err = l.RedisClient.HSet(false, idCacheKey, id, token); err != nil {
		return "", fmt.Errorf("save credential error: %s", err)
	}

	return token, nil
}

// Guard 访问守卫
// token: 访问凭证
// 返回值: 续期后的新token（如果有续期），错误信息
func (l *Gatekeeper) Guard(token string) (string, error) {
	if token == "" {
		return "", errors.New("empty token")
	}

	// 解密token
	data, err := l.decrypt(token)
	if err != nil {
		return "", fmt.Errorf("decrypt error: %s", err)
	}

	// 检查黑名单
	blacklisted, blacklistItem, err := l.IsBlacklisted(data.ID)
	if err != nil {
		functions.LogError("check blacklist error:", err)
	} else if blacklisted {
		return "", fmt.Errorf("access denied: id %s is blacklisted, reason: %s", data.ID, blacklistItem.Reason)
	}

	// 限流检查
	if allowed, current, err := l.RateLimit(data.ID); err != nil {
		functions.LogError("rate limit error:", err)
	} else if !allowed {
		_, _, resetTime, _ := l.GetRateLimitStatus(data.ID)
		return "", fmt.Errorf(
			"rate limit exceeded: current %d, limit %d, reset at %s",
			current,
			l.rateLimit,
			resetTime.Format("15:04:05"),
		)
	}

	// 检查凭证是否存在
	cacheToken, err := l.RedisClient.HGet(idCacheKey, data.ID)
	if err != nil {
		if redis.CheckNil(err) {
			return "", errors.New("credential not found")
		}

		return "", fmt.Errorf("redis error: %s", err)
	}

	if token != string(cacheToken) {
		return "", errors.New("invalid token")
	}

	var now = uint64(time.Now().UnixMilli())
	// 检查是否过期
	if now > data.Expire {
		// 异步清理过期凭证
		go l.cleanupExpiredToken(data.ID)
		return "", errors.New("token has expired")
	}

	// 剩余时间
	var remainingTime = data.Expire - now
	// 如果剩余时间小于总有效期的一半，自动续期
	if remainingTime < l.Expire/2 {
		newToken, err := l.Credential(data.ID)
		if err != nil {
			functions.LogError("renew token error:", err)
			return "", nil // 返回空字符串表示不需要续期
		}
		return newToken, nil
	}

	return "", nil
}

// cleanupExpiredToken 异步清理过期凭证
func (l *Gatekeeper) cleanupExpiredToken(id string) {
	select {
	case clearLockChan <- struct{}{}:
		defer func() { <-clearLockChan }()

		// 再次检查是否真的过期
		cacheToken, err := l.RedisClient.HGet(idCacheKey, id)
		if err != nil {
			return
		}

		data, err := l.decrypt(string(cacheToken))
		if err != nil {
			return
		}

		if uint64(time.Now().UnixMilli()) > data.Expire {
			_, _ = l.RedisClient.HDel(idCacheKey, id)
			_, _ = l.RedisClient.SRem(uniqueIdsCacheKey, id)
		}
	default:
		// 无法获取锁，跳过清理
	}
}

// encryption 加密缓存项生成token
func (l *Gatekeeper) encryption(ipt *CacheItem) (string, error) {
	if ipt == nil {
		return "", errors.New("cache item cannot be nil")
	}

	b, err := functions.JSONMarshal(ipt)
	if err != nil {
		return "", fmt.Errorf("marshal error: %s", err)
	}

	encrypt, err := functions.NewCrypto([]byte(l.Key)).Encrypt(b)
	if err != nil {
		return "", fmt.Errorf("encrypt error: %s", err)
	}

	return prefix + l.swapStringParts(encrypt) + suffix, nil
}

// decrypt 解密token获取缓存项
func (l *Gatekeeper) decrypt(ipt string) (*CacheItem, error) {
	if ipt == "" {
		return nil, errors.New("token cannot be empty")
	}

	ipt = strings.TrimSpace(ipt)
	ipt = strings.TrimPrefix(ipt, prefix)
	ipt = strings.TrimSuffix(ipt, suffix)

	if ipt == "" {
		return nil, errors.New("invalid token format")
	}

	decrypted, err := functions.NewCrypto([]byte(l.Key)).Decrypt(l.swapStringParts(ipt))
	if err != nil {
		return nil, fmt.Errorf("decrypt error: %s", err)
	}

	var data CacheItem
	if err = functions.JSONUnmarshal([]byte(decrypted), &data); err != nil {
		return nil, fmt.Errorf("unmarshal error: %s", err)
	}

	if data.ID == "" {
		return nil, errors.New("invalid cache item: missing id")
	}

	return &data, nil
}

// swapStringParts 字符串部分交换，用于简单的混淆
func (l *Gatekeeper) swapStringParts(s string) string {
	if len(s) < 10 {
		return s
	}

	return s[len(s)-3:] + s[3:len(s)-3] + s[:3]
}
