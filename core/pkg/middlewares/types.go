package middlewares

import (
	"context"
	"net/http"

	jwt "github.com/golang-jwt/jwt/v4"

	"skeyevss/core/localization"
	"skeyevss/core/tps"
)

type (
	MailCall     func() func(info, broken string)
	CustomerCall func(ctx context.Context, r *http.Request) (context.Context, *localization.Item)

	Conf struct {
		AesKey, // 对撑加密
		Secret string // jwt key
		Expire, // 过期时间
		TokenType int64 // TOKEN_TYPE_MULTIPART TOKEN_TYPE_SINGLE
		TokenVerification bool // 校验token

		MailCall
		CustomerCall
	}

	MW struct {
		Expire int64 // 失效时间
		Encrypt,
		Key string // 秘钥
		TokenType int64 // TOKEN_TYPE_MULTIPART TOKEN_TYPE_SINGLE

		MailFunc     MailCall
		CustomerFunc CustomerCall

		CustomParse func(claims jwt.MapClaims) (tps.MI, error)
	}
)
