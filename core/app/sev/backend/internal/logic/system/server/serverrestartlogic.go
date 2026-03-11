package server

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/common"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/pkg/response"
)

type ServerRestartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewServerRestartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ServerRestartLogic {
	return &ServerRestartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ServerRestartLogic) ServerRestart() *response.HttpErr {
	return common.Restart(l.svcCtx, "")
}
