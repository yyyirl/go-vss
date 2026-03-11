package server

import (
	"net/http"

	"skeyevss/core/app/sev/backend/internal/logic/system/server"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/pkg/response"
)

func SystemInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()
		response.New().Success(ctx, w, server.NewSystemInfoLogic(ctx, svcCtx).SystemInfo())
	}
}
