package ms

import (
	"net/http"

	"skeyevss/core/app/sev/backend/internal/logic/config/ms"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/pkg/response"
)

func OptionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()
		resp, err := ms.NewOptionsLogic(ctx, svcCtx).Options()
		if err != nil {
			response.New().RequestError(ctx, w, err)
			return
		}

		response.New().Success(ctx, w, resp)
	}
}
