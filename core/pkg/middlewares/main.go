package middlewares

import (
	"context"
	"errors"
	"net/http"
	"time"

	"skeyevss/core/constants"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

var tokenExpire = errors.New("token expired")

func New(conf Conf, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx          = r.Context()
			xProtocol    = r.Header.Get(constants.HEADER_XProtocol)     // https 请求 解决nginx代理识别为http
			platform     = r.Header.Get(constants.HEADER_PLATFORM)      // 客户端类型
			language     = r.Header.Get(constants.HEADER_LANG)          // 语言
			token        = r.Header.Get(constants.HEADER_AUTHORIZATION) // authorization
			mac          = r.Header.Get(constants.HEADER_MAC)           // authorization
			refreshToken = r.FormValue(constants.HEADER_REFRESH_TOKEN)  // refresh authorization
		)

		ip, _ := functions.GetIP(r)
		// recover 通知
		defer (&MW{MailFunc: conf.MailCall}).recover(w, r)

		ctx = context.WithValue(ctx, constants.CTX_REQUESTS, map[string]interface{}{
			"ip":         ip,
			"remoteAddr": r.RemoteAddr,
			"url":        r.Host + r.RequestURI,
			"referer":    r.Referer(),
			"method":     r.Method,
			"header":     r.Header,
		})
		ctx = context.WithValue(ctx, constants.HEADER_IP, ip)
		ctx = context.WithValue(ctx, constants.HEADER_MAC, mac)
		ctx = context.WithValue(ctx, constants.HEADER_LANG, language)
		ctx = context.WithValue(ctx, constants.HEADER_PLATFORM, platform)
		ctx = context.WithValue(ctx, constants.CTX_REQ_START_TIME, time.Now().UnixMilli())

		if r.TLS != nil || xProtocol == "1" {
			ctx = context.WithValue(ctx, constants.HEADER_TLS, "https://"+r.Host)
		} else {
			ctx = context.WithValue(ctx, constants.HEADER_TLS, "http://"+r.Host)
		}

		if conf.AesKey == "" {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Str("系统参数`aes key`错误"), localization.M0004))
			return
		}

		if conf.Expire <= 0 {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Str("系统参数`expire`错误"), localization.M0005))
			return
		}

		tokenInfo, newToken, err := (&MW{
			Key:       conf.AesKey,
			Encrypt:   token,
			Expire:    conf.Expire,
			TokenType: conf.TokenType,
		}).Vase(refreshToken != "")
		if err != nil {
			if conf.TokenVerification {
				if errors.Is(err, tokenExpire) {
					response.New().RequestError(ctx, w, response.MakeUnauthorizedError(err, localization.M0006))
					return
				}

				response.New().RequestError(ctx, w, response.MakeUnauthorizedError(err, localization.M0006))
				return
			}
		}

		if tokenInfo != nil && tokenInfo.Userinfo != nil {
			ctx = context.WithValue(ctx, constants.CTX_USERID, tokenInfo.Userinfo["id"])
		}

		// 登录信息
		if tokenInfo != nil {
			ctx = context.WithValue(ctx, constants.CTX_TOKEN_INFO, tokenInfo)
		}
		if newToken != "" {
			ctx = context.WithValue(ctx, constants.HEADER_NEW_TOKEN, newToken)
		}

		// 自定义验证
		if conf.CustomerCall != nil {
			var err *localization.Item
			ctx, err = conf.CustomerCall(ctx, r)
			if err != nil {
				response.New().RequestError(ctx, w, response.MakeForbiddenError(errors.New("自定义验证失败"), err))
				return
			}
		}

		next(w, r.WithContext(ctx))
	}
}
