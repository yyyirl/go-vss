package gbs_proc

import (
	"time"

	"skeyevss/core/app/sev/vss/internal/types"
)

var _ types.SipProcLogic = (*HeartbeatOfflineLogic)(nil)

type HeartbeatOfflineLogic struct {
	svcCtx      *types.ServiceContext
	recoverCall func(name string)
}

// 心跳

func (l *HeartbeatOfflineLogic) DO(params *types.DOProcLogicParams) {
	l = &HeartbeatOfflineLogic{
		svcCtx:      params.SvcCtx,
		recoverCall: params.RecoverCall,
	}
	l.svcCtx.InitFetchDataState.Wait()

	defer l.recoverCall("心跳检测消息接收")

	// 创建定时器
	go l.proc()

	for {
		select {
		case v := <-l.svcCtx.SipHeartbeatLoop:
			l.svcCtx.SipHeartbeatLoopMap.Set(v.ID, v)
		}
	}
}

func (l *HeartbeatOfflineLogic) proc() {
	defer l.recoverCall("心跳检测 loop")

	for val := range time.NewTicker(time.Second * 1).C {
		// functions.LogInfo("定时器 heartbeat_loop l.svcCtx.SipCatalogLoopMap.Values: ", l.svcCtx.SipHeartbeatLoopMap.Len())
		for _, item := range l.svcCtx.SipHeartbeatLoopMap.Values() {
			var now = val.Unix()
			// 注册失效 登录超时检测 || 心跳超时
			if now-item.RegisterExpireAt > 10 || now-item.Now >= l.svcCtx.Config.Sip.HeartbeatTimeout {
				// 更新设备状态下线
				l.svcCtx.SipHeartbeatLoopMap.Remove(item.ID)
				l.svcCtx.SetDeviceOnline <- &types.DCOnlineReq{
					DeviceUniqueId: item.ID,
					Online:         false,
				}
			}
		}
	}
}
