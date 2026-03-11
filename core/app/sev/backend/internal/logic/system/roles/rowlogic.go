package roles

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/roles"
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

func (l *RowLogic) Row(req *types.IdQuery) (interface{}, *response.HttpErr) {
	res, err := response.NewRpcToHttpResp[*backendservice.Response, *roles.Item]().Parse(
		func() (*backendservice.Response, error) {
			return l.svcCtx.RpcClients.Backend.RoleRow(l.ctx, &backendservice.IDReq{ID: uint64(req.Id)})
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}
