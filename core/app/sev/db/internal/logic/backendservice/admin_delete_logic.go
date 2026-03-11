package backendservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/admins"
)

type AdminDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDeleteLogic {
	return &AdminDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminDeleteLogic) AdminDelete(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
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
		&orm.ConditionItem{
			Column:   admins.ColumnSuper,
			Operator: "<",
			Value:    1,
		},
	)

	return nil, response.NewMakeRpcRetErr(l.svcCtx.AdminsModel.DeleteBy(params), 2)
}
