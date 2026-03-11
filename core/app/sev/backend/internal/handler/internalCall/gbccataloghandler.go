package internalCall

import (
	"net/http"

	"skeyevss/core/app/sev/backend/internal/logic/internalCall"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/common/source/permissions"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/common"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/response"
)

func GBCCatalogHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()
		if err := permissions.New(ctx).Authentication(contextx.GetSuperState(ctx), permissions.P_0_7_1_9, contextx.GetPermissionIds(ctx)); err != nil {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M1006))
			return
		}

		var req map[string]interface{}
		if err := common.Parse(r, &req); err != nil {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001))
			return
		}

		if err := internalCall.NewGBCCatalogLogic(ctx, svcCtx).GBCCatalog(req); err != nil {
			response.New().RequestError(ctx, w, err)
			return
		}

		response.New().Success(ctx, w, nil)
	}
}
