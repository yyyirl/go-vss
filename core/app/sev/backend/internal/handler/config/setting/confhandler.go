package setting

import (
	"net/http"

	"skeyevss/core/app/sev/backend/internal/logic/config/setting"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/pkg/response"
)

func ConfHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()
		response.New().Success(ctx, w, setting.NewConfLogic(ctx, svcCtx).Conf())
	}
}
