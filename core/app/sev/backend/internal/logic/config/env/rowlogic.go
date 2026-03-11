package env

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type RowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RowLogic {
	return &RowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RowLogic) Row() (interface{}, *response.HttpErr) {
	content, err := functions.ReadFile(l.svcCtx.Config.EnvFile)
	if err != nil {
		return 0, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00277)
	}

	return map[string]interface{}{
		"content": string(content),
		"file":    l.svcCtx.Config.EnvFile,
	}, nil
}
