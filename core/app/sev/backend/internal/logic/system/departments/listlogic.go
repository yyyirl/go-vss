package departments

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
	"skeyevss/core/repositories/models/departments"
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
	res, err := response.NewRpcToHttpResp[*backendservice.Response, *response.ListWithMapResp[[]*departments.Item, uint64]]().Parse(
		func() (*backendservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(req)
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Backend.Departments(l.ctx, data)
		},
	)
	if err != nil {
		return nil, err
	}

	var depIds []interface{}
	for _, item := range res.Data.List {
		depIds = append(depIds, item.ID)
	}

	if len(depIds) > 0 {
		// 获取通道列表信息
		channelRes, err1 := response.NewRpcToHttpResp[*deviceservice.Response, []*channels.Item]().Parse(
			func() (*deviceservice.Response, error) {
				data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
					Conditions: []*orm.ConditionItem{
						{
							Column:   channels.ColumnDepIds,
							Original: orm.NewExternalDB(l.svcCtx.Config.SevBase.DatabaseType).MakeCaseJSONContainsCondition(channels.ColumnDepIds, depIds),
						},
					},
					All:            true,
					IgnoreNotFound: true,
				})
				if err != nil {
					return nil, err
				}

				return l.svcCtx.RpcClients.Device.ChannelXlist(l.ctx, data)
			},
		)
		if err1 != nil {
			return nil, err1
		}

		var maps = make(map[uint64][]*channels.Item)
		for _, item := range channelRes.Data {
			for _, id := range item.DepIds {
				_, ok := maps[id]
				if !ok {
					maps[id] = make([]*channels.Item, 0)
				}

				maps[id] = append(maps[id], item)
			}
		}

		res.Data.Maps = functions.MapToMapInterface(
			functions.MapArrUniqueWithCall(maps, func(item *channels.Item) string {
				return fmt.Sprintf("%d-%s", item.ID, item.UniqueId)
			}),
		)
	}

	return res.Data, nil
}
