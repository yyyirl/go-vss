// @Title        gatekeeper
// @Description  main
// @Create       yiyiyi 2025/9/26 13:44

package gatekeeper

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/redis"
)

var (
	clearExpireToken    sync.Once
	clearLockChan       = make(chan struct{}, 1)
	concurrenceLockChan = make(chan struct{}, 1)
)

func New(redisClient *redis.GoRedisClient, expire uint64, key, node string) *Gatekeeper {
	var instance = &Gatekeeper{
		RedisClient: redisClient,
		Key:         key,
		Node:        node,
		Expire:      expire,
	}

	clearExpireToken.Do(func() {
		go instance.clearExpireToken()
	})
	return instance
}

func (l *Gatekeeper) clearExpireToken() {
	var tick = time.Tick(10 * time.Second)
	for {
		select {
		case v := <-tick:
			l.doClearExpireToken(v)
		}
	}
}

func (l *Gatekeeper) doClearExpireToken(v time.Time) {
	clearLockChan <- struct{}{}
	defer func() { <-clearLockChan }()

	maps, err := l.RedisClient.HGetAll(idCacheKey)
	if err != nil {
		functions.LogError("fetch all expire token error:", err)
		return
	}

	var deleteIds []string
	for key, item := range maps {
		data, err := l.decrypt(item)
		if err != nil {
			deleteIds = append(deleteIds, key)
			continue
		}

		if uint64(v.UnixMilli()) > data.Expire {
			deleteIds = append(deleteIds, key)
			deleteIds = append(deleteIds, data.ID)
		}
	}

	deleteIds = functions.ArrUnique(deleteIds)
	if len(deleteIds) <= 0 {
		return
	}

	// 删除过期token
	if _, err := l.RedisClient.HDel(idCacheKey, deleteIds...); err != nil {
		functions.LogError("delete expire token error:", err)
	}

	if _, err := l.RedisClient.SRem(idCacheKey, functions.SliceToSliceAny(deleteIds)...); err != nil {
		functions.LogError("delete expire id error:", err)
	}
}

func (l *Gatekeeper) ClientID() (string, error) {
	return l.clientID(0, "")
}

func (l *Gatekeeper) clientID(counter int, desc string) (string, error) {
	if counter >= 100 {
		return "", fmt.Errorf("retry counter out of range[%s]", desc)
	}

	// ok, err := l.RedisClient.AcquireLock(produceIDRedisLockKey, l.Node, 1*time.Second)
	// if err != nil {
	// 	return "", err
	// }
	//
	// if !ok {
	// 	concurrenceLockChan <- struct{}{}
	// 	<-concurrenceLockChan
	//
	// 	return l.clientID(counter+1, "concurrence. retry")
	// }
	//
	// defer func() {
	// 	if _, err := l.RedisClient.ReleaseLock(produceIDRedisLockKey, l.Node); err != nil {
	// 		functions.LogError("release expire token lock error:", err)
	// 		return
	// 	}
	// }()

	var id = strconv.FormatInt(time.Now().UnixNano(), 10)
	if exists, _ := l.RedisClient.SIsMember(uniqueIdsCacheKey, id); exists {
		concurrenceLockChan <- struct{}{}
		<-concurrenceLockChan

		return l.clientID(counter+1, "already exists. retry")
	}

	if _, err := l.RedisClient.SAdd(uniqueIdsCacheKey, id); err != nil {
		functions.LogError("create unique cache id error:", err)
	}

	return id, nil
}

// 生成凭证
func (l *Gatekeeper) Credential(id string) (string, error) {
	clearLockChan <- struct{}{}
	<-clearLockChan

	token, err := l.encryption(&CacheItem{ID: id, Expire: uint64(functions.NewTimer().NowMilli()) + l.Expire})
	if err != nil {
		return "", err
	}

	if _, err := l.RedisClient.HSet(false, idCacheKey, id, token); err != nil {
		return "", err
	}

	return token, nil
}

// 访问守卫
func (l *Gatekeeper) Guard(token string) (string, error) {
	if token == "" {
		return "", errors.New("empty token")
	}

	data, err := l.decrypt(token)
	if err != nil {
		return "", err
	}

	cacheToken, err := l.RedisClient.HGet(idCacheKey, data.ID)
	if err != nil {
		if redis.CheckNil(err) {
			return "", errors.New("Nonexistent Credential")
		}

		return "", err
	}

	if token != string(cacheToken) {
		return "", errors.New("The token is invalid.")
	}

	var now = uint64(functions.NewTimer().NowMilli())
	if data.Expire > now {
		return "", errors.New("The token has expired.")
	}

	var remainingTime = math.Abs(float64(data.Expire - now))
	if remainingTime >= float64(l.Expire)/2 {
		// 续期token
		return l.Credential(data.ID)
	}

	// TODO 黑名单
	// TODO 限流

	return "", nil
}

func (l *Gatekeeper) encryption(ipt *CacheItem) (string, error) {
	if ipt == nil {
		return "", errors.New("激活信息不能为空")
	}

	b, err := functions.JSONMarshal(ipt)
	if err != nil {
		return "", err
	}

	encrypt, err := functions.NewCrypto([]byte(l.Key)).Encrypt(b)
	if err != nil {
		return "", err
	}

	return prefix + l.swapStringParts(encrypt) + suffix, nil
}

func (l *Gatekeeper) decrypt(ipt string) (*CacheItem, error) {
	if ipt == "" {
		return nil, errors.New("秘钥不能为空")
	}

	ipt = strings.TrimSpace(ipt)
	ipt = strings.TrimPrefix(ipt, prefix)
	ipt = strings.TrimSuffix(ipt, suffix)

	b, err := functions.NewCrypto([]byte(l.Key)).Decrypt(l.swapStringParts(ipt))
	if err != nil {
		return nil, err
	}

	var data CacheItem
	if err := functions.JSONUnmarshal([]byte(b), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (l *Gatekeeper) swapStringParts(s string) string {
	if len(s) < 10 {
		return s
	}

	return s[len(s)-3:] + s[3:len(s)-3] + s[:3]
}
