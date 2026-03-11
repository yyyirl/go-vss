package pkg

import (
	"errors"
	"strings"

	"skeyevss/core/app/sev/vss/internal/config"
	"skeyevss/core/pkg/functions"
)

type (
	UserTokenData struct {
		ID       uint64 `json:"id"`       // 用户id
		Expire   uint64 `json:"expire"`   // 过期时间
		ConnType string `json:"connType"` // 链接类型
	}

	UserToken struct {
		Data   *UserTokenData
		config config.Config
	}
)

var wsFlags = map[string]string{
	"+": "*",
	"/": "-",
	"=": "!",
}

func NewAes(conf config.Config) *UserToken {
	return &UserToken{
		config: conf,
	}
}

func (u *UserToken) MakeUserToken(data *UserTokenData) (string, error) {
	if data == nil {
		return "", errors.New("data is nil")
	}

	data.Expire = uint64(functions.NewTimer().NowMilli() + u.config.XAuth.LoginExpire)
	b, err := functions.JSONMarshal(data)
	if err != nil {
		return "", err
	}

	encrypt, err := functions.NewCrypto([]byte(u.config.XAuth.AesKey)).Encrypt(b)
	if err != nil {
		return "", err
	}

	for k, v := range wsFlags {
		encrypt = strings.ReplaceAll(encrypt, k, v)
	}

	return encrypt, nil
}

func (u *UserToken) ParseUserToken(data string) (*UserTokenData, error) {
	for k, v := range wsFlags {
		data = strings.ReplaceAll(data, v, k)
	}

	if data == "" {
		return nil, errors.New("data is nil")
	}

	content, err := functions.NewCrypto([]byte(u.config.XAuth.AesKey)).Decrypt(
		strings.Replace(data, " ", "+", -1),
	)
	if err != nil {
		return nil, err
	}

	var res UserTokenData
	if err := functions.JSONUnmarshal([]byte(content), &res); err != nil {
		return nil, err
	}

	if u.config.XAuth.ConnType != res.ConnType {
		return nil, errors.New("connType error")
	}

	if uint64(functions.NewTimer().Now()) >= res.Expire {
		return nil, errors.New("token expired")
	}

	return &res, nil
}

func (u *UserToken) MakeXAuthorization(id uint64) (string, error) {
	authorization, err := NewUtils(
		&UtilsConfig{
			Expire: uint64(functions.NewTimer().Now() + u.config.XAuth.LoginExpire/1000),
		},
	).MakeConnAuthorization(u.config.XAuth.AesKey, u.config.XAuth.ConnType, id)
	if err != nil {
		return "", err
	}

	for k, v := range wsFlags {
		authorization = strings.ReplaceAll(authorization, k, v)
	}

	return authorization, nil
}

func (u *UserToken) ParseXAuthorization(data string) error {
	for k, v := range wsFlags {
		data = strings.ReplaceAll(data, v, k)
	}

	return NewUtils(nil).VerifyConnAuthorization(u.config.XAuth.AesKey, u.config.XAuth.ConnType, data)
}
