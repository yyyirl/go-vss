package deviceservicelogic

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/common/types"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
)

type OnlineStateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOnlineStateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OnlineStateLogic {
	return &OnlineStateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设备在线状态
func (l *OnlineStateLogic) OnlineState(_ *db.XRequestParams) (*db.Response, error) {
	// 设备
	deviceList, err := l.svcCtx.DevicesModel.OnlineStateList(l.ctx, new(orm.ReqParams))
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	var deviceMaps = make(map[string]uint)
	for _, item := range deviceList {
		deviceMaps[item.DeviceUniqueId] = item.Online
	}

	// 通道
	channelList, err := l.svcCtx.ChannelsModel.OnlineStateList(l.ctx, new(orm.ReqParams))
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	var channelMaps = make(map[string]uint)
	for _, item := range channelList {
		channelMaps[fmt.Sprintf("%s-%s", item.DeviceUniqueId, item.UniqueId)] = item.Online
	}

	return response.NewRpcResp[*db.Response]().Make(&types.DeviceOnlineStateResp{
		Channels: channelMaps,
		Devices:  deviceMaps,
	}, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
