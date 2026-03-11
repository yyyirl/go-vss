package configservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
	"skeyevss/core/repositories/models/departments"
)

type MaxIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMaxIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MaxIdLogic {
	return &MaxIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取uniqueId
func (l *MaxIdLogic) MaxId(in *db.TypeReq) (*db.Response, error) {
	var id string
	if in.Type == "cascadeDepCode" {
		if err := l.svcCtx.DepartmentsModel.MaxWithParams(
			departments.ColumnCascadeDepUniqueId,
			&id,
			&orm.ReqParams{
				IgnoreNotFound: true,
				Conditions: []*orm.ConditionItem{
					{
						Column:   departments.ColumnCascadeDepUniqueId,
						Original: l.svcCtx.ChannelsModel.CaseNumberCondition(departments.ColumnCascadeDepUniqueId),
					},
				},
			},
		); err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}
	} else if in.Type == "cascadeChannel" {
		if err := l.svcCtx.ChannelsModel.MaxWithParams(
			channels.ColumnCascadeChannelUniqueId,
			&id,
			&orm.ReqParams{
				IgnoreNotFound: true,
				Conditions: []*orm.ConditionItem{
					{
						Column:   channels.ColumnCascadeChannelUniqueId,
						Original: l.svcCtx.ChannelsModel.CaseNumberCondition(channels.ColumnCascadeChannelUniqueId),
					},
				},
			},
		); err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}
	}

	return response.NewRpcResp[*db.Response]().Make(id, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
