package gbs_sip

import (
	"context"

	gosip "github.com/ghettovoice/gosip/sip"

	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/app/sev/vss/internal/pkg/sip"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/devices"
)

var _ types.SipReceiveHandleLogic[*KeepaliveLogic] = (*KeepaliveLogic)(nil)

type KeepaliveLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	req    *types.Request
	tx     gosip.ServerTransaction
}

func (l *KeepaliveLogic) New(ctx context.Context, svcCtx *types.ServiceContext, req *types.Request, tx gosip.ServerTransaction) *KeepaliveLogic {
	return &KeepaliveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		req:    req,
		tx:     tx,
	}
}

func (l *KeepaliveLogic) DO() *types.Response {
	// 心跳带有鉴权信息 ----------------------------------------------------------------
	if len(l.req.ID) < 18 {
		return &types.Response{
			Error: types.NewErr("参数错误"),
			Code:  types.StatusUnauthorized,
		}
	}
	// if len(l.req.Authorization) > 0 {
	// 	// 不应答 让设备超时重新注册
	// 	return &types.Response{
	// 		Error: types.NewErr(types.DisableResponseError.Error()),
	// 		Code:  types.StatusPreconditionFailed,
	// 	}
	// }

	// 心跳数据解析
	data, err := sip.NewParser[types.SipMessageKeepalive]().ToData(l.req.Original)
	if err != nil {
		return &types.Response{Error: types.NewErr(err.Error())}
	}

	// 没有鉴权信息 ----------------------------------------------------------------

	// 获取设备信息
	deviceRes, err1 := response.NewRpcToHttpResp[*backendservice.Response, *devices.Item]().Parse(
		func() (*backendservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: devices.ColumnDeviceUniqueId, Value: data.DeviceID},
				},
				IgnoreNotFound: true,
			})
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.DeviceRow(l.ctx, data)
		},
	)
	if err1 != nil {
		return &types.Response{Error: types.NewErr(err1.Error)}
	}

	if deviceRes.Data == nil {
		return nil
	}

	var now = functions.NewTimer().Now()
	// 设置上线状态
	if deviceRes.Data.Online == 0 {
		if deviceRes.Data.Expire <= uint64(now)+5 {
			// 注册已过期
			return &types.Response{
				Error: types.NewErr(types.DisableResponseError.Error()),
				Code:  types.StatusPreconditionFailed,
			}
		}

		// 注册未过期重新上线 !!! 有一定概率会和注册下线冲突 等待下次 注册/心跳任务 更正状态 心跳和注册同时进入
	}

	l.svcCtx.SetDeviceOnline <- &types.DCOnlineReq{
		DeviceUniqueId: deviceRes.Data.DeviceUniqueId,
		Online:         true,
	}

	// 已在线 服务重启后没有任务
	if _, ok := l.svcCtx.SipCatalogLoopMap.Get(l.req.ID); !ok {
		// 立即发送catalog请求
		l.req.Caller = functions.CallerFile(2)
		l.svcCtx.SipSendCatalog <- l.req
		// 注册定时发送catalog任务
		l.svcCtx.SipCatalogLoop <- &types.SipCatalogLoopReq{
			Req:    l.req,
			Online: true,
			Now:    now,
		}
	}

	// 设置更新检测心跳
	if v, ok := l.svcCtx.SipHeartbeatLoopMap.Get(l.req.ID); ok {
		l.svcCtx.SipHeartbeatLoop <- &types.SipHeartbeatLoopReq{
			ID:               data.DeviceID,
			Now:              now,
			RegisterExpireAt: v.RegisterExpireAt,
		}
	} else {
		l.svcCtx.SipHeartbeatLoop <- &types.SipHeartbeatLoopReq{
			ID:               data.DeviceID,
			Now:              now,
			RegisterExpireAt: int64(deviceRes.Data.Expire),
		}
	}

	return nil
}
