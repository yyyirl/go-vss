package items

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/common"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/departments"
	"skeyevss/core/repositories/models/devices"
)

type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListLogic) List(req *orm.ReqParams) (interface{}, *response.HttpErr) {
	var departmentIds = common.DepartmentIds(l.ctx)
	if departmentIds != nil {
		req.Conditions = append(req.Conditions, &orm.ConditionItem{
			Column:   devices.ColumnDepIds,
			Original: orm.NewExternalDB(l.svcCtx.Config.SevBase.DatabaseType).MakeCaseJSONContainsCondition(devices.ColumnDepIds, functions.SliceToSliceAny(departmentIds)),
		})
	}

	res, err := response.NewRpcToHttpResp[*deviceservice.Response, response.ListWithExtResp[[]*devices.Item, []string, uint64]]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(req)
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.DeviceList(l.ctx, data)
		},
	)
	if err != nil {
		return nil, err
	}

	var depIds []uint64
	for _, item := range res.Data.List {
		depIds = append(depIds, item.DepIds...)
	}

	if len(depIds) > 0 {
		// 获取部门信息
		departmentRes, err := response.NewRpcToHttpResp[*backendservice.Response, *response.ListResp[[]*departments.Item]]().Parse(
			func() (*backendservice.Response, error) {
				data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
					Conditions: []*orm.ConditionItem{
						{Column: departments.ColumnId, Values: functions.SliceToSliceAny(depIds)},
					},
					IgnoreNotFound: true,
					Limit:          len(depIds),
				})
				if err != nil {
					return nil, err
				}

				return l.svcCtx.RpcClients.Backend.Departments(l.ctx, data)
			},
		)
		if err != nil {
			return nil, err
		}

		if len(departmentRes.Data.List) > 0 {
			res.Data.Ext = make(map[uint64]interface{})
			for _, item := range departmentRes.Data.List {
				res.Data.Ext[item.ID] = item.Name
			}
		}
	}

	return res.Data, nil
}
