package middleware

import (
	"context"
	"net/http"

	"skeyevss/core/app/sev/backend/internal/config"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/middlewares"
)

type BaseMiddleware struct {
}

func NewBaseMiddleware() *BaseMiddleware {
	return &BaseMiddleware{}
}

func (m *BaseMiddleware) Handle(c config.Config, next http.HandlerFunc, buildTime string) http.HandlerFunc {
	return middlewares.New(
		middlewares.Conf{
			AesKey: c.Auth.AesKey,
			Secret: c.Auth.JwtSecret,
			Expire: c.Auth.LoginExpire,
			MailCall: func() func(info, broken string) {
				return recoverCallback(c)
			},
			CustomerCall: func(ctx context.Context, r *http.Request) (context.Context, *localization.Item) {
				return ctx, nil
			},
		},
		next,
	)
}
