package server

import (
	"net/http"

	"skeyevss/core/app/sev/backend/internal/logic/system/server"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/pkg/response"
)

func HealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response.New().Success(r.Context(), w, server.NewHealthLogic(r.Context(), svcCtx).Health())
	}
}
