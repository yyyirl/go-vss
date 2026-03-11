package internalCall

import (
	"net/http"

	"skeyevss/core/app/sev/backend/internal/logic/internalCall"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/common/source/permissions"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/response"
)

func OnvifDiscoverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()
		if err := permissions.New(ctx).Authentication(contextx.GetSuperState(ctx), permissions.P_0_7_1_12, contextx.GetPermissionIds(ctx)); err != nil {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M1006))
			return
		}

		resp, err := internalCall.NewOnvifDiscoverLogic(ctx, svcCtx).OnvifDiscover()
		if err != nil {
			response.New().RequestError(ctx, w, err)
			return
		}

		response.New().Success(ctx, w, resp)
	}
}
