package gbs_proc

import (
	"context"
	"time"

	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
	"skeyevss/core/repositories/models/devices"
)

var _ types.SipProcLogic = (*SetDeviceOnlineStateLogic)(nil)

type SetDeviceOnlineStateLogic struct {
	svcCtx      *types.ServiceContext
	recoverCall func(name string)
}

// 定时发送catalog
func (l *SetDeviceOnlineStateLogic) DO(params *types.DOProcLogicParams) {
	l = &SetDeviceOnlineStateLogic{
		svcCtx:      params.SvcCtx,
		recoverCall: params.RecoverCall,
	}
	l.svcCtx.InitFetchDataState.Wait()

	defer l.recoverCall("心跳检测消息接收")

	// 创建定时器
	go l.proc()

	for {
		select {
		case v := <-l.svcCtx.SetDeviceOnline:
			l.svcCtx.DeviceOnlineStateUpdateMap.Set(v.DeviceUniqueId, v)
		}
	}
}

func (l *SetDeviceOnlineStateLogic) proc() {
	defer l.recoverCall("设备在线状态检测 loop")

	for range time.NewTicker(time.Second).C {
		var (
			channelOnline,
			channelOffline []uint64

			deviceOnline,
			deviceOffline []string
		)
		for _, item := range l.svcCtx.DeviceOnlineStateUpdateMap.Values() {
			if item.ChannelUniqueId != "" {
				if item.Online {
					channelOnline = append(channelOnline, item.CId)
				} else {
					channelOffline = append(channelOffline, item.CId)
				}
				continue
			}
			if item.Online {
				deviceOnline = append(deviceOnline, item.DeviceUniqueId)
			} else {
				deviceOffline = append(deviceOffline, item.DeviceUniqueId)
			}
		}

		if len(deviceOnline) > 0 {
			go l.setDevice(deviceOnline, 1)
		}

		if len(deviceOffline) > 0 {
			go l.setDevice(deviceOffline, 0)
		}

		if len(channelOnline) > 0 {
			go l.setChannel(channelOnline, 1)
		}

		if len(channelOffline) > 0 {
			go l.setChannel(channelOffline, 0)
		}

		l.svcCtx.DeviceOnlineStateUpdateMap.Clear()
	}
}

// 更新设备上线状态
func (l *SetDeviceOnlineStateLogic) setDevice(ids []string, online uint) {
	var (
		now     = functions.NewTimer().NowMilli()
		records = []*orm.UpdateItem{
			{Column: devices.ColumnOnline, Value: online},
			{Column: devices.ColumnKeepaliveAt, Value: now},
		}
	)
	// 下线停止catalog
	if online == 0 {
		for _, item := range ids {
			l.svcCtx.SipCatalogLoop <- &types.SipCatalogLoopReq{
				Req: &types.Request{
					ID: item,
				},
				Online: false,
				Now:    functions.NewTimer().Now(),
			}
		}

		records = append(records, &orm.UpdateItem{Column: devices.ColumnOfflineAt, Value: now})

		if _, err := response.NewRpcToHttpResp[*deviceservice.Response, uint64]().Parse(
			func() (*deviceservice.Response, error) {
				data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
					Conditions: []*orm.ConditionItem{
						{
							Column: devices.ColumnDeviceUniqueId,
							Values: functions.SliceToSliceAny(ids),
						},
					},
					Data: []*orm.UpdateItem{
						{Column: channels.ColumnOnline, Value: online},
					},
				})
				if err != nil {
					return nil, err
				}

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				return l.svcCtx.RpcClients.Device.ChannelUpdate(ctx, data)
			},
		); err != nil {
			functions.LogError("设备通道在线状态更新失败, err: ", err.Error)
		}
	} else {
		records = append(records, &orm.UpdateItem{Column: devices.ColumnOnlineAt, Value: now})
	}

	if _, err := response.NewRpcToHttpResp[*deviceservice.Response, uint64]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{
						Column: devices.ColumnDeviceUniqueId,
						Values: functions.SliceToSliceAny(ids),
					},
				},
				Data: records,
			})
			if err != nil {
				return nil, err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			return l.svcCtx.RpcClients.Device.DeviceUpdate(ctx, data)
		},
	); err != nil {
		functions.LogError("设备在线状态更新失败, err: ", err.Error)
	}
}

func (l *SetDeviceOnlineStateLogic) setChannel(ids []uint64, online uint) {
	if _, err := response.NewRpcToHttpResp[*deviceservice.Response, uint64]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{
						Column: devices.ColumnID,
						Values: functions.SliceToSliceAny(ids),
					},
				},
				Data: []*orm.UpdateItem{
					{Column: channels.ColumnOnline, Value: online},
				},
			})
			if err != nil {
				return nil, err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			return l.svcCtx.RpcClients.Device.ChannelUpdate(ctx, data)
		},
	); err != nil {
		functions.LogError("设备通道在线状态更新失败[1], err: ", err.Error)
	}
}
