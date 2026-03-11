package deviceservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
	"skeyevss/core/repositories/models/devices"
)

type ChannelListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChannelListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChannelListLogic {
	return &ChannelListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChannelListLogic) ChannelList(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	// 获取总数
	count, queryErr := l.svcCtx.ChannelsModel.Count(params)
	if queryErr != nil {
		return nil, response.NewMakeRpcRetErr(queryErr, 2)
	}

	if count <= 0 {
		return response.NewRpcResp[*db.Response]().Make(response.NewListResp[[]*channels.Item]().Empty(), 3, func(data []byte) *db.Response {
			return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
		})
	}

	// 获取列表
	list, queryErr := l.svcCtx.ChannelsModel.List(params)
	if queryErr != nil {
		return nil, response.NewMakeRpcRetErr(queryErr, 2)
	}

	var (
		records         []*channels.Item
		deviceUniqueIds []string
	)
	for _, item := range list {
		v, err := item.ConvToItem()
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		records = append(records, v)
		deviceUniqueIds = append(deviceUniqueIds, item.DeviceUniqueId)
	}

	// 获取设备信息
	var deviceMaps = make(map[string]interface{})
	if len(deviceUniqueIds) > 0 {
		list, err := l.svcCtx.DevicesModel.List(
			&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{
						Column: devices.ColumnDeviceUniqueId,
						Values: functions.SliceToSliceAny(deviceUniqueIds),
					},
				},
				All: true,
			},
		)
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		for _, item := range list {
			v, err := item.ConvToItem()
			if err != nil {
				return nil, response.NewMakeRpcRetErr(err, 2)
			}

			deviceMaps[item.DeviceUniqueId] = v
		}
	}

	return response.NewRpcResp[*db.Response]().Make(&response.ListWithMapResp[[]*channels.Item, string]{
		List:  records,
		Count: count,
		Maps:  deviceMaps,
	}, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
