package gbs

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/app/sev/vss/internal/pkg/ms"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/common/stream"
	"skeyevss/core/constants"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/devices"
)

var (
	_ types.HttpEHandleLogic[*stopStreamLogic] = (*stopStreamLogic)(nil)

	StopStreamLogic = new(stopStreamLogic)
)

type stopStreamLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *stopStreamLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *stopStreamLogic {
	return &stopStreamLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *stopStreamLogic) Path() string {
	return "/gbs/stop-stream-keepalive/:streamName/:retry" // retry 1 重新拉取invite
}

func (l *stopStreamLogic) MakePath(streamName string, retry uint) string {
	return fmt.Sprintf("/api/gbs/stop-stream-keepalive/%s/%d", streamName, retry)
}

// 发送 bye请求 stop接口 ms停止拉流
func (l *stopStreamLogic) DO() *types.HttpResponse {
	return l.StopStream(l.c.Param("streamName"), l.c.Param("retry"))
}

func (l *stopStreamLogic) StopStream(streamName, retry string) *types.HttpResponse {
	res, err := stream.New().Parse(streamName)
	if err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.MR1003),
		}
	}

	_, ok := l.svcCtx.SipCatalogLoopMap.Get(res.Device)
	if !ok {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(constants.DeviceUnregistered), localization.M00300),
		}
	}

	// 获取设备信息
	deviceRes, err1 := response.NewRpcToHttpResp[*backendservice.Response, *devices.Item]().Parse(
		func() (*backendservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: devices.ColumnDeviceUniqueId, Value: res.Device},
				},
			})
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.DeviceRow(l.ctx, data)
		},
	)
	if err1 != nil {
		return &types.HttpResponse{Err: err1}
	}

	var msNode = ms.New(l.ctx, l.svcCtx).VoteNode(deviceRes.Data.MSIds)

	// 发送bye 请求
	l.svcCtx.SipSendBye <- &types.SipByeMessage{
		Data:       res,
		StreamName: streamName,
	}

	// 停止ms stream
	if err := ms.New(l.ctx, l.svcCtx).StopMSStream(msNode.Address, streamName, "pub"); err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.MR1003),
		}
	}

	// // 停止任务
	// if _, err := redis.NewStreamKeepaliveRunningState(l.svcCtx.RedisClient).Set(streamName, false); err != nil {
	// 	return &types.HttpResponse{Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.MR1003)}
	// }

	// 重新发起invite请求
	if retry != "0" {
		// 发送保活心跳
		if res := InviteLogic.New(l.ctx, l.c, l.svcCtx).Invite(&InviteParams{
			DeviceUniqueId: res.Device,
			ChannelID:      res.Channel,
			PlayType:       res.PlayType,
			Caller:         "http 请求stream stop重试invite",
		}); res != nil && res.Err != nil {
			return res
		}
	}

	return nil
}
