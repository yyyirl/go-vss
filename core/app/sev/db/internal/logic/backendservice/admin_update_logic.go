package backendservicelogic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/admins"
)

type AdminUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateLogic {
	return &AdminUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminUpdateLogic) AdminUpdate(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	if v, ok := params.DataRecord[admins.ColumnPassword]; ok {
		p, ok := v.(string)
		if !ok {
			return nil, response.NewMakeRpcRetErr(errors.New("password type is invalid"), 2)
		}

		if p == "" || functions.IsBcryptHash(p) {
			delete(params.DataRecord, admins.ColumnPassword)
		}
	}

	record, err := admins.NewItem().CheckMap(params.DataRecord)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	params.Conditions = append(
		params.Conditions,
		&orm.ConditionItem{
			Column:   admins.ColumnId,
			Operator: "NOTIN",
			Values:   []interface{}{1, 2},
		},
	)

	return nil, response.NewMakeRpcRetErr(l.svcCtx.AdminsModel.UpdateWithParams(record, params), 2)
}
