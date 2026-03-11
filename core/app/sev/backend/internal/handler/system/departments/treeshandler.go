package departments

import (
	"net/http"

	"skeyevss/core/app/sev/backend/internal/logic/system/departments"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/common/source/permissions"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/response"
)

func TreesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()
		if err := permissions.New(ctx).Authentication(contextx.GetSuperState(ctx), permissions.P_0_1_3_6, contextx.GetPermissionIds(ctx)); err != nil {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M1006))
			return
		}

		resp, err := departments.NewTreesLogic(ctx, svcCtx).Trees()
		if err != nil {
			response.New().RequestError(ctx, w, err)
			return
		}

		response.New().Success(ctx, w, resp)
	}
}
