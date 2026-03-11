/**
 * @Author:         yi
 * @Description:    对称加密校验
 * @Version:        1.0.0
 * @Date:           2022/10/11 11:20
 */
package middlewares

import (
	"errors"
	"strings"
	"time"

	"github.com/allegro/bigcache"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/tps"
)

// 刷新token
const refreshTokenExpireReturn = 30

var tokenRefreshCache *bigcache.BigCache

func init() {
	var err error
	tokenRefreshCache, err = functions.BigCacheThrottle(refreshTokenExpireReturn * time.Second)
	if err != nil {
		panic(err)
	}
}

func (m *MW) Vase(refresh bool) (*tps.TokenItem, string, error) {
	if m.Key == "" {
		return nil, "", errors.New("`key` param is necessary")
	}

	if m.Encrypt == "" {
		return nil, "", errors.New("`encrypt` param is necessary")
	}

	m.Encrypt = strings.Replace(m.Encrypt, " ", "+", -1)
	encrypt, err := functions.NewCrypto([]byte(m.Key)).Decrypt(m.Encrypt)
	if err != nil {
		return nil, "", err
	}

	var res tps.TokenItem
	if err := functions.JSONUnmarshal([]byte(encrypt), &res); err != nil {
		return nil, "", err
	}

	var expire = res.Expire - functions.NewTimer().NowMilli()
	if expire <= 0 {
		return nil, "", tokenExpire
	}

	if m.Expire/2 >= expire || refresh {
		uniqueId, err := functions.ToUniqueId(res)
		if err != nil {
			return nil, "", err
		}

		// 生成新token
		if _, err := tokenRefreshCache.Get(uniqueId); err == bigcache.ErrEntryNotFound {
			encrypt, err := MakeTokenVASE(m.Key, m.Expire, res)
			if err != nil {
				return nil, "", err
			}

			if err := tokenRefreshCache.Set(uniqueId, []byte("1")); err != nil {
				return nil, "", err
			}

			return &res, encrypt, nil
		}
	}

	return &res, "", nil
}

func MakeTokenVASE(key string, expire int64, data tps.TokenItem) (string, error) {
	data.Expire = functions.NewTimer().Now() + expire
	b, err := functions.JSONMarshal(data)
	if err != nil {
		return "", err
	}

	encrypt, err := functions.NewCrypto([]byte(key)).Encrypt(b)
	if err != nil {
		return "", err
	}

	return encrypt, nil
}
