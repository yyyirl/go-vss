package deviceservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/common/types"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
	"skeyevss/core/repositories/models/devices"
)

type DeviceChannelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeviceChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeviceChannelLogic {
	return &DeviceChannelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设备列表
func (l *DeviceChannelLogic) DeviceChannel(in *db.DeviceChannelReq) (*db.Response, error) {
	deviceRow, err := l.svcCtx.DevicesModel.RowWithParams(&orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: devices.ColumnDeviceUniqueId, Value: in.DeviceUniqueId},
		},
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	deviceItem, err := deviceRow.ConvToItem()
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	channelRow, err := l.svcCtx.ChannelsModel.RowWithParams(&orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: channels.ColumnUniqueId, Value: in.ChannelUniqueId},
			{Column: channels.ColumnDeviceUniqueId, Value: in.DeviceUniqueId},
		},
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	channelItem, err := channelRow.ConvToItem()
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return response.NewRpcResp[*db.Response]().Make(&types.DeviceChannel{
		Device:  deviceItem,
		Channel: channelItem,
	}, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
