package internalCall

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"skeyevss/core/app/sev/backend/internal/logic/internalCall"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/common/source/permissions"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/response"
)

func MSQueryRecordByNamesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()
		if err := permissions.New(ctx).Authentication(contextx.GetSuperState(ctx), permissions.P_0_7_1_2, contextx.GetPermissionIds(ctx)); err != nil {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M1006))
			return
		}

		var req types.MsQueryRecordByNameReq
		if err := httpx.Parse(r, &req); err != nil {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001))
			return
		}

		resp, err := internalCall.NewMSQueryRecordByNamesLogic(ctx, svcCtx).MSQueryRecordByNames(&req)
		if err != nil {
			response.New().RequestError(ctx, w, err)
			return
		}

		response.New().Success(ctx, w, resp)
	}
}
