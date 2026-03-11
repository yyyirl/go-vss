package server

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"skeyevss/core/app/sev/backend/internal/svc"

	"skeyevss/core/pkg/response"
)

type SystemUpdatesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSystemUpdatesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SystemUpdatesLogic {
	return &SystemUpdatesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SystemUpdatesLogic) SystemUpdates() (interface{}, *response.HttpErr) {
	return nil, nil
}
