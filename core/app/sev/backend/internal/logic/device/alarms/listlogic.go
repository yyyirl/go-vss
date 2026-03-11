package alarms

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	ctypes "skeyevss/core/common/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/alarms"
	"skeyevss/core/repositories/models/channels"
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
	res, err := response.NewRpcToHttpResp[*deviceservice.Response, *response.ListWithMapResp[[]*alarms.Item, string]]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(req)
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.AlarmList(l.ctx, data)
		},
	)
	if err != nil {
		return nil, err
	}

	var channelUniqueIds []string
	for _, item := range res.Data.List {
		channelUniqueIds = append(channelUniqueIds, item.DeviceUniqueId)
	}

	res.Data.Ext = make(map[string]interface{})
	channelUniqueIds = functions.ArrUnique(channelUniqueIds)
	if len(channelUniqueIds) > 0 {
		data, err := response.NewRpcToHttpResp[*deviceservice.Response, *ctypes.DeviceChannels]().Parse(
			func() (*deviceservice.Response, error) {
				return l.svcCtx.RpcClients.Device.DeviceChannelRelationsWithChannelIds(l.ctx, &deviceservice.UniqueIdsReq{
					UniqueIds: channelUniqueIds,
				})
			},
		)
		if err != nil {
			return nil, err
		}

		{
			var maps = make(map[string]map[string]interface{})
			for _, item := range data.Data.Channels {
				maps[item.UniqueId] = map[string]interface{}{
					channels.ColumnName:           item.Name,
					channels.ColumnLabel:          item.Label,
					channels.ColumnDeviceUniqueId: item.DeviceUniqueId,
				}
			}
			res.Data.Ext["channels"] = maps
		}

		{
			var maps = make(map[string]map[string]interface{})
			for _, item := range data.Data.Devices {
				maps[item.DeviceUniqueId] = map[string]interface{}{
					devices.ColumnName:  item.Name,
					devices.ColumnLabel: item.Label,
				}
			}
			res.Data.Ext["devices"] = maps
		}
	}

	return res.Data, nil
}
