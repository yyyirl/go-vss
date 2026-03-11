package channels

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
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
	res, err := response.NewRpcToHttpResp[*deviceservice.Response, *channels.Item]().Parse(
		func() (*deviceservice.Response, error) {
			return l.svcCtx.RpcClients.Device.ChannelRow(l.ctx, &deviceservice.IDReq{ID: uint64(req.Id)})
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}
