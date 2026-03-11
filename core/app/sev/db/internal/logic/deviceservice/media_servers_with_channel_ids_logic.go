package deviceservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	cTypes "skeyevss/core/common/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/devices"
)

type MediaServersWithChannelIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMediaServersWithChannelIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MediaServersWithChannelIdsLogic {
	return &MediaServersWithChannelIdsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MediaServersWithChannelIdsLogic) MediaServersWithChannelIds(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	// 获取列表
	channelList, err := l.svcCtx.ChannelsModel.XList1(l.ctx, params)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	var deviceUniqueIds []string
	for _, item := range channelList {
		deviceUniqueIds = append(deviceUniqueIds, item.DeviceUniqueId)
	}

	deviceMaps, err := l.svcCtx.DevicesModel.MSMaps(l.ctx, &orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: devices.ColumnDeviceUniqueId, Values: functions.SliceToSliceAny(deviceUniqueIds)},
		},
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	var data = make(map[uint64]*cTypes.ChannelMSRelItem)
	for _, item := range channelList {
		data[item.ID] = &cTypes.ChannelMSRelItem{
			ChannelId:       item.ID,
			ChannelUniqueId: item.UniqueId,
			DeviceUniqueId:  item.DeviceUniqueId,
			MSIds:           deviceMaps[item.DeviceUniqueId],
		}
	}

	return response.NewRpcResp[*db.Response]().Make(data, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
