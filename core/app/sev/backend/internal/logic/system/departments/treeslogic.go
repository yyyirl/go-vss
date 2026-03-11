package departments

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/common"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/categories"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/departments"
)

type TreesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTreesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TreesLogic {
	return &TreesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TreesLogic) Trees() (interface{}, *response.HttpErr) {
	var departmentIds = common.DepartmentIds(l.ctx)
	if departmentIds != nil {
		if len(departmentIds) <= 0 {
			return 0, response.MakeError(response.NewHttpRespMessage().Str("未分配组织部门"), localization.MR1006)
		}
	}

	res, err := response.NewRpcToHttpResp[*backendservice.Response, []*categories.Item[int, *departments.Item]]().Parse(
		func() (*backendservice.Response, error) {
			return l.svcCtx.RpcClients.Backend.DepartmentTrees(l.ctx, &backendservice.IdsReq{Ids: departmentIds})
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}
