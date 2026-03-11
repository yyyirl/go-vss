package gbs_proc

import (
	"context"
	"time"

	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/vss/internal/pkg/ms"
	"skeyevss/core/app/sev/vss/internal/types"
	cTypes "skeyevss/core/common/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/devices"
)

var _ types.SipProcLogic = (*CheckDeviceOnlineStateLogic)(nil)

type CheckDeviceOnlineStateLogic struct {
	executing bool

	svcCtx      *types.ServiceContext
	recoverCall func(name string)
}

func (l *CheckDeviceOnlineStateLogic) DO(params *types.DOProcLogicParams) {
	l = &CheckDeviceOnlineStateLogic{
		svcCtx:      params.SvcCtx,
		recoverCall: params.RecoverCall,
	}
	l.svcCtx.InitFetchDataState.Wait()

	defer l.recoverCall("拉流状态检测")
	go l.proc()
}

func (l *CheckDeviceOnlineStateLogic) proc() {
	defer l.recoverCall("拉流状态检测 loop")

	l.executing = true
	go l.check()

	// 3s执行一次更新
	for range time.NewTicker(time.Second * 30).C {
		if l.executing {
			continue
		}

		l.executing = true
		go l.check()
	}
}

// 更新设备上线状态
func (l *CheckDeviceOnlineStateLogic) check() {
	defer func() {
		l.executing = false
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := response.NewRpcToHttpResp[*deviceservice.Response, *cTypes.RtspStreamGroupResp]().Parse(
		func() (*deviceservice.Response, error) {
			return l.svcCtx.RpcClients.Device.RtspStreamGroups(ctx, &deviceservice.IdsReq{
				Ids: []uint64{uint64(devices.AccessProtocol_1), uint64(devices.AccessProtocol_2), uint64(devices.AccessProtocol_3)},
			})
		},
	)
	if err != nil {
		functions.LogError("rtsp 分组获取失败, err:", err)
		return
	}

	if res.Data == nil {
		functions.LogInfo("------------------------------没有可更新数据")
		return
	}

	for _, item := range append(res.Data.Onvif, res.Data.StreamSourceRtsp...) {
		go func(item *cTypes.RtspStreamGroupItem) {
			var online = false
			if item.StreamUrl != "" {
				if res, _ := ms.NewRTSPChecker(5 * time.Second).Check(item.StreamUrl); res != nil && res.IsOnline {
					online = true
				}
			}

			l.svcCtx.SetDeviceOnline <- &types.DCOnlineReq{
				DeviceUniqueId:  item.DeviceUniqueId,
				ChannelUniqueId: item.ChannelUniqueId,
				CId:             item.CId,
				Online:          online,
			}
		}(item)
	}

	for _, item := range res.Data.Rtmp {
		go func(item *cTypes.RtspStreamGroupItem) {
			var online = false
			if item.StreamUrl != "" {
				if res, _ := ms.NewRTMPChecker(5*time.Second, false).Check(item.StreamUrl); res != nil && res.IsOnline {
					online = true
				}
			}

			l.svcCtx.SetDeviceOnline <- &types.DCOnlineReq{
				DeviceUniqueId:  item.DeviceUniqueId,
				ChannelUniqueId: item.ChannelUniqueId,
				CId:             item.CId,
				Online:          online,
			}
		}(item)
	}

	for _, item := range res.Data.Http {
		go func(item *cTypes.RtspStreamGroupItem) {
			var online = false
			if item.StreamUrl != "" {
				if res, _ := ms.NewHTTPChecker(5 * time.Second).CheckStream(item.StreamUrl); res != nil && res.IsOnline {
					online = true
				}
			}

			l.svcCtx.SetDeviceOnline <- &types.DCOnlineReq{
				DeviceUniqueId:  item.DeviceUniqueId,
				ChannelUniqueId: item.ChannelUniqueId,
				CId:             item.CId,
				Online:          online,
			}
		}(item)
	}
}

// 设备
// l.svcCtx.SetDeviceOnline <- &types.DCOnlineReq{
// 	DeviceUniqueId: item.DeviceUniqueId,
// 	Online:         res.IsOnline,
//
// }

// 流媒体源 默认在线

// RTMP推流
// 通过 on_pub_start.go 设置在线状态
// 通过 on_pub_stop.go 设置离线

// ONVIF协议
// 通过设备发现获取状态 存在则在线

// GB28181协议
// 检测心跳 keepalive
