package deviceservicelogic

import (
	"context"
	"strings"

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

type RtspStreamGroupsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRtspStreamGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RtspStreamGroupsLogic {
	return &RtspStreamGroupsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// rtsp stream url 分组 accessProtocol
func (l *RtspStreamGroupsLogic) RtspStreamGroups(in *db.IdsReq) (*db.Response, error) {
	// 设备
	deviceList, err := l.svcCtx.DevicesModel.SList(l.ctx, &orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: devices.ColumnAccessProtocol, Values: functions.SliceToSliceAny(in.Ids)},
		},
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	if len(deviceList) <= 0 {
		return response.NewRpcResp[*db.Response]().Make(nil, 3, func(data []byte) *db.Response {
			return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
		})
	}

	type (
		CItem struct {
			ID        uint64
			UniqueId  string
			StreamUrl string
		}

		DItem struct {
			AccessProtocol uint
			StreamUrl      string
			Channels       []*CItem
		}
	)

	var (
		deviceMaps      = make(map[string]*DItem)
		deviceUniqueIds []string
	)
	for _, item := range deviceList {
		deviceUniqueIds = append(deviceUniqueIds, item.DeviceUniqueId)

		deviceMaps[item.DeviceUniqueId] = &DItem{
			StreamUrl:      item.StreamUrl,
			AccessProtocol: item.AccessProtocol,
		}
	}

	// 通道
	channelList, err := l.svcCtx.ChannelsModel.List(&orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: channels.ColumnDeviceUniqueId, Values: functions.SliceToSliceAny(deviceUniqueIds)},
		},
		All: true,
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	for _, item := range channelList {
		if _, ok := deviceMaps[item.DeviceUniqueId]; ok {
			deviceMaps[item.DeviceUniqueId].Channels = append(
				deviceMaps[item.DeviceUniqueId].Channels,
				&CItem{
					ID:        item.ID,
					UniqueId:  item.UniqueId,
					StreamUrl: item.StreamUrl,
				},
			)
		}
	}

	var resp = new(types.RtspStreamGroupResp)
	for key, item := range deviceMaps {
		// 流媒体源
		if item.AccessProtocol == devices.AccessProtocol_1 {
			for _, v := range item.Channels {
				if strings.Index(item.StreamUrl, "http") == 0 {
					resp.Http = append(
						resp.Http,
						&types.RtspStreamGroupItem{
							CId:             v.ID,
							ChannelUniqueId: v.UniqueId,
							DeviceUniqueId:  key,
							StreamUrl:       item.StreamUrl,
						},
					)
				} else if strings.Index(item.StreamUrl, "rtmp") == 0 {
					resp.Rtmp = append(
						resp.Rtmp,
						&types.RtspStreamGroupItem{
							CId:             v.ID,
							ChannelUniqueId: v.UniqueId,
							DeviceUniqueId:  key,
							StreamUrl:       v.StreamUrl,
						},
					)
				} else {
					resp.StreamSourceRtsp = append(
						resp.StreamSourceRtsp,
						&types.RtspStreamGroupItem{
							CId:             v.ID,
							ChannelUniqueId: v.UniqueId,
							DeviceUniqueId:  key,
							StreamUrl:       item.StreamUrl,
						},
					)
				}
			}
			continue
		}

		// onvif
		if item.AccessProtocol == devices.AccessProtocol_3 {
			for _, v := range item.Channels {
				resp.Onvif = append(
					resp.Onvif,
					&types.RtspStreamGroupItem{
						CId:             v.ID,
						ChannelUniqueId: v.UniqueId,
						DeviceUniqueId:  key,
						StreamUrl:       v.StreamUrl,
					},
				)
			}
		}

		// rtmp推流模式下通过on-pub-start/stop判断在线状态，不通过拉流的方式检测流在线状态
		if item.AccessProtocol == devices.AccessProtocol_2 {
			// for _, v := range item.Channels {
			// 	resp.Rtmp = append(
			// 		resp.Rtmp,
			// 		&types.RtspStreamGroupItem{
			// 			CId:             v.ID,
			// 			ChannelUniqueId: v.UniqueId,
			// 			DeviceUniqueId:  key,
			// 			StreamUrl:       v.StreamUrl,
			// 		},
			// 	)
			// }
		}
	}

	return response.NewRpcResp[*db.Response]().Make(resp, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
