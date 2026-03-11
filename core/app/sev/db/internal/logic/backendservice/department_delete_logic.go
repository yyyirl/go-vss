package backendservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/response"
)

type DepartmentDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDepartmentDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DepartmentDeleteLogic {
	return &DepartmentDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DepartmentDeleteLogic) DepartmentDelete(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return nil, response.NewMakeRpcRetErr(l.svcCtx.DepartmentsModel.DeleteBy(params), 2)
}
