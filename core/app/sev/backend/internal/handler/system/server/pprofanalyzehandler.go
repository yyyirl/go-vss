package server

import (
	"net/http"

	v "github.com/go-playground/validator/v10"

	"skeyevss/core/app/sev/backend/internal/logic/system/server"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/common/source/permissions"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/common"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/pprof"
	"skeyevss/core/pkg/response"
)

func PProfAnalyzeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()
		if err := permissions.New(ctx).Authentication(contextx.GetSuperState(ctx), permissions.P_0_6_5, contextx.GetPermissionIds(ctx)); err != nil {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M1006))
			return
		}

		var req pprof.PProfParams
		if err := common.Parse(r, &req); err != nil {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001))
			return
		}

		if err := v.New().Struct(req); err != nil {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001))
			return
		}

		resp, err := server.NewPProfAnalyzeLogic(ctx, svcCtx).PProfAnalyze(&req)
		if err != nil {
			response.New().RequestError(ctx, w, err)
			return
		}

		response.New().Success(ctx, w, resp)
	}
}
