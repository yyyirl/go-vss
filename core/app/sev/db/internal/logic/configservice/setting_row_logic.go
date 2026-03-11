package configservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/settings"
)

type SettingRowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSettingRowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SettingRowLogic {
	return &SettingRowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SettingRowLogic) SettingRow(_ *db.EmptyRequest) (*db.Response, error) {
	row, err := l.svcCtx.SettingModel.RowWithParams(&orm.ReqParams{
		Orders: []*orm.OrderItem{
			{Column: settings.ColumnId, Value: orm.SORT_DESC},
		},
		Conditions: []*orm.ConditionItem{
			{Column: settings.ColumnId, Value: 0, Operator: ">"},
		},
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	data, err := row.ConvToItem()
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return response.NewRpcResp[*db.Response]().Make(data, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
