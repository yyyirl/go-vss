package deviceservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	cTypes "skeyevss/core/common/types"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
	"skeyevss/core/repositories/models/devices"
)

type DeviceOnlineStatisticsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeviceOnlineStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeviceOnlineStatisticsLogic {
	return &DeviceOnlineStatisticsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设备在线统计
func (l *DeviceOnlineStatisticsLogic) DeviceOnlineStatistics(_ *db.EmptyRequest) (*db.Response, error) {
	// 在线通道数量
	channelOnlineCount, err := l.svcCtx.ChannelsModel.Count(&orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: channels.ColumnOnline, Value: 1},
		},
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	// 离线通道数量
	channelOfflineCount, err := l.svcCtx.ChannelsModel.Count(&orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: channels.ColumnOnline, Value: 0},
		},
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	// 在线通道数量
	deviceOnlineCount, err := l.svcCtx.DevicesModel.Count(&orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: devices.ColumnOnline, Value: 1},
		},
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	// 离线通道数量
	deviceOfflineCount, err := l.svcCtx.DevicesModel.Count(&orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: devices.ColumnOnline, Value: 0},
		},
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	// 接入类型统计
	accessProtocolGroup, err := l.svcCtx.DevicesModel.GroupByAccessProtocol()
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return response.NewRpcResp[*db.Response]().Make(&cTypes.DeviceStatisticsResp{
		ChannelOnlineCount:  channelOnlineCount,
		ChannelOfflineCount: channelOfflineCount,
		DeviceOnlineCount:   deviceOnlineCount,
		DeviceOfflineCount:  deviceOfflineCount,
		AccessProtocolGroup: accessProtocolGroup,
	}, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
