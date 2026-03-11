package {{.pkgName}}

import (
	{{.imports}}
	{{if eq .request `req *types.FindParams`}}"skeyevss/core/pkg/orm"{{end}}
	 "skeyevss/core/pkg/response"
)

type {{.logic}} struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func New{{.logic}}(ctx context.Context, svcCtx *svc.ServiceContext) *{{.logic}} {
	return &{{.logic}}{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}


func (l *{{.logic}}) {{.function}}({{if eq .request `req *types.FindParams`}}req *orm.FindParams{{else}}{{.request}}{{end}}) ({{if eq .responseType `error`}}interface{}{{else}}{{.responseType}}{{end}}, *response.HttpErr) {
	// todo: add your logic here and delete this line

	return nil, nil
}
