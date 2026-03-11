package ms

import (
	"context"
	"skeyevss/core/repositories/models/media-servers"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/app/sev/db/client/configservice"
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

func (l *RowLogic) Row(req *types.IdQuery) (interface{}, *response.HttpErr) {
	res, err := response.NewRpcToHttpResp[*configservice.Response, *mediaServers.Item]().Parse(
		func() (*configservice.Response, error) {
			return l.svcCtx.RpcClients.Config.MsRow(l.ctx, &configservice.IDReq{ID: uint64(req.Id)})
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}
