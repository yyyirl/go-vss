package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/admins"
)

type InfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InfoLogic {
	return &InfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InfoLogic) Info() (interface{}, *response.HttpErr) {
	res, err := response.NewRpcToHttpResp[*backendservice.Response, *admins.Item]().Parse(
		func() (*backendservice.Response, error) {
			return l.svcCtx.RpcClients.Backend.AdminRow(l.ctx, &backendservice.IDReq{ID: uint64(contextx.GetCtxUserid(l.ctx))})
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}
