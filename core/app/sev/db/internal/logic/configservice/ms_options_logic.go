package configservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	mediaServers "skeyevss/core/repositories/models/media-servers"
	"skeyevss/core/tps"
)

type MsOptionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMsOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MsOptionsLogic {
	return &MsOptionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// media server option 列表
func (l *MsOptionsLogic) MsOptions(_ *db.EmptyRequest) (*db.Response, error) {
	// 获取列表
	list, queryErr := l.svcCtx.MediaServersModel.List(
		&orm.ReqParams{
			All: true,
			Orders: []*orm.OrderItem{
				{Column: mediaServers.ColumnCreatedAt, Value: orm.SORT_DESC},
			},
			Conditions: []*orm.ConditionItem{
				{Column: mediaServers.ColumnIsDef, Value: 0},
			},
		},
	)
	if queryErr != nil {
		return nil, response.NewMakeRpcRetErr(queryErr, 2)
	}

	var records []*tps.OptionItem
	for _, item := range list {
		records = append(records, &tps.OptionItem{
			Title: item.Name,
			Value: item.ID,
			Raw:   item,
		})
	}

	return response.NewRpcResp[*db.Response]().Make(records, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
