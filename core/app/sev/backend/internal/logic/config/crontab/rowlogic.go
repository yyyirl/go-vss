package crontab

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/app/sev/db/client/configservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/crontab"
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

func (l *RowLogic) Row(req *types.UniqueIdQuery) (interface{}, *response.HttpErr) {
	res, err := response.NewRpcToHttpResp[*configservice.Response, *crontab.Item]().Parse(
		func() (*configservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: crontab.ColumnUniqueId, Value: req.UniqueId},
				},
			})
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Config.CrontabRow(l.ctx, data)
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}
