package deviceservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/common/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
	"skeyevss/core/repositories/models/devices"
)

type DeviceChannelRelationsWithChannelIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeviceChannelRelationsWithChannelIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeviceChannelRelationsWithChannelIdsLogic {
	return &DeviceChannelRelationsWithChannelIdsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设备通道列表(通过通道id获取)
func (l *DeviceChannelRelationsWithChannelIdsLogic) DeviceChannelRelationsWithChannelIds(in *db.UniqueIdsReq) (*db.Response, error) {
	var (
		channelRecords   []*channels.Item
		deviceRecords    []*devices.Item
		deviceUniqueIds  []string
		channelUniqueIds = functions.ArrUnique(in.UniqueIds)
	)
	if len(channelUniqueIds) > 0 {
		// 获取通道
		records, err := l.svcCtx.ChannelsModel.List(&orm.ReqParams{
			Conditions: []*orm.ConditionItem{
				{Column: channels.ColumnUniqueId, Values: functions.SliceToSliceAny(channelUniqueIds)},
			},
			Limit: len(channelUniqueIds),
		})
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		for _, item := range records {
			v, err := item.ConvToItem()
			if err != nil {
				return nil, response.NewMakeRpcRetErr(err, 2)
			}

			channelRecords = append(channelRecords, v)
			deviceUniqueIds = append(deviceUniqueIds, item.DeviceUniqueId)
		}
	}

	deviceUniqueIds = functions.ArrUnique(deviceUniqueIds)
	if len(deviceUniqueIds) > 0 {
		// 获取设备
		records, err := l.svcCtx.DevicesModel.List(&orm.ReqParams{
			Conditions: []*orm.ConditionItem{
				{Column: devices.ColumnDeviceUniqueId, Values: functions.SliceToSliceAny(deviceUniqueIds)},
			},
			Limit: len(deviceUniqueIds),
		})
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		for _, item := range records {
			v, err := item.ConvToItem()
			if err != nil {
				return nil, response.NewMakeRpcRetErr(err, 2)
			}

			deviceRecords = append(deviceRecords, v)
		}
	}

	return response.NewRpcResp[*db.Response]().Make(&types.DeviceChannels{
		Devices:  deviceRecords,
		Channels: channelRecords,
	}, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
