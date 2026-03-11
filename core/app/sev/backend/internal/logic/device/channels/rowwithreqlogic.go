package channels

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
)

type RowWithReqLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRowWithReqLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RowWithReqLogic {
	return &RowWithReqLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RowWithReqLogic) RowWithReq(req *orm.ReqParams) (interface{}, *response.HttpErr) {
	res, err := response.NewRpcToHttpResp[*deviceservice.Response, *channels.Item]().Parse(
		func() (*deviceservice.Response, error) {
			params, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(req)
			if err != nil {
				return nil, response.NewMakeRpcRetErr(err, 2)
			}

			return l.svcCtx.RpcClients.Device.ChannelRowFind(l.ctx, params)
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}
