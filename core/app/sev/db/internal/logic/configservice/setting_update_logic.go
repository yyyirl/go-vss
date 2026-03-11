package configservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/settings"
)

type SettingUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSettingUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SettingUpdateLogic {
	return &SettingUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SettingUpdateLogic) SettingUpdate(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	record, err := settings.NewItem().CheckMap(params.DataRecord)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	params.Conditions = []*orm.ConditionItem{{Column: settings.ColumnId, Value: 1}}
	return nil, response.NewMakeRpcRetErr(l.svcCtx.SettingModel.UpdateWithParams(record, params), 2)
}
