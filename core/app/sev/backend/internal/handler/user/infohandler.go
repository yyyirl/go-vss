package user

import (
	"net/http"

	"skeyevss/core/app/sev/backend/internal/logic/user"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/pkg/response"
)

func InfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()
		resp, err := user.NewInfoLogic(ctx, svcCtx).Info()
		if err != nil {
			response.New().RequestError(ctx, w, err)
			return
		}

		response.New().Success(ctx, w, resp)
	}
}
