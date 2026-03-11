package backendservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/departments"
)

type DepartmentUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDepartmentUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DepartmentUpdateLogic {
	return &DepartmentUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DepartmentUpdateLogic) DepartmentUpdate(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	record, err := departments.NewItem().CheckMap(params.DataRecord)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return nil, response.NewMakeRpcRetErr(l.svcCtx.DepartmentsModel.UpdateWithParams(record, params), 2)
}
