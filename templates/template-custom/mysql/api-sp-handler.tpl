package {{.PkgName}}

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	 "skeyevss/core/pkg/response"
	 "skeyevss/core/localization"
	{{if .HasRequest}}{{if eq .RequestType `FindParams`}}"skeyevss/core/pkg/orm"{{end}}{{end}}
	{{.ImportPackages}}
)

func {{.HandlerName}}(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		{{if .HasRequest}}{{if eq .RequestType `FindParams`}}var (
			req orm.FindParams
			ctx = r.Context()
		)
		if err := httpx.Parse(r, &req); err != nil {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001))
			return
		}{{else}}var (
			ctx = r.Context()
			req types.{{.RequestType}}
		)
		if err := httpx.Parse(r, &req); err != nil {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001))
			return
		}{{end}}

		resp, err := {{.LogicName}}.New{{.LogicType}}(ctx, svcCtx).{{.Call}}(&req);
		if err != nil {
			response.New().RequestError(ctx, w, err)
			return
		}{{else}}var ctx = r.Context()
		resp, err := {{.LogicName}}.New{{.LogicType}}(ctx, svcCtx).{{.Call}}();
		if err != nil {
			response.New().RequestError(ctx, w, err)
            return
        }{{end}}

		response.New().Success(ctx, w, resp)
	}
}