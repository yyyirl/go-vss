package login

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"skeyevss/core/app/sev/backend/internal/logic/login"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx = r.Context()
			req types.LoginReq
		)
		if err := httpx.Parse(r, &req); err != nil {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001))
			return
		}

		resp, err := login.NewLoginLogic(ctx, svcCtx).Login(&req)
		if err != nil {
			response.New().RequestError(ctx, w, err)
			return
		}

		response.New().Success(ctx, w, resp)
	}
}
