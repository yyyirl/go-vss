package backendservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/categories"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/departments"
)

type DepartmentTreesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDepartmentTreesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DepartmentTreesLogic {
	return &DepartmentTreesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DepartmentTreesLogic) DepartmentTrees(req *db.IdsReq) (*db.Response, error) {
	var conditions []*orm.ConditionItem
	if len(req.Ids) > 0 {
		conditions = append(conditions, &orm.ConditionItem{
			Column: departments.ColumnId,
			Values: functions.SliceToSliceAny(req.Ids),
		})
	}

	list, queryErr := l.svcCtx.DepartmentsModel.List(&orm.ReqParams{
		All:        true,
		Conditions: conditions,
	})
	if queryErr != nil {
		return nil, response.NewMakeRpcRetErr(queryErr, 2)
	}

	var records []*departments.Item
	for _, item := range list {
		v, err := item.ConvToItem()
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		records = append(records, v)
	}

	return response.NewRpcResp[*db.Response]().Make(
		categories.New[int, *departments.Item]().Conv(
			records,
			func(item *departments.Item) *categories.Item[int, *departments.Item] {
				return &categories.Item[int, *departments.Item]{
					ID: int(item.ID), Pid: int(item.ParentId), Name: item.Name, Raw: item,
				}
			},
		).Trees,
		3,
		func(data []byte) *db.Response {
			return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
		},
	)
}
