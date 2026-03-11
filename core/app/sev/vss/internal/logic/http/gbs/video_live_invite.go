package gbs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/app/sev/vss/internal/pkg/ms"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/common/stream"
	"skeyevss/core/constants"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
	"skeyevss/core/repositories/models/devices"
)

var (
	_ types.HttpEHandleLogic[*inviteLogic] = (*inviteLogic)(nil)

	InviteLogic = new(inviteLogic)
)

type inviteLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

type InviteParams struct {
	DeviceUniqueId string
	ChannelID      string
	PlayType       stream.PlayType
	StartAt        string
	EndAt          string
	Download       bool
	Speed          float64
	StreamName     string

	Caller string

	KeepaliveCron bool
	OnPubStart    bool

	ChannelItem *channels.Item
	DeviceItem  *devices.Item
}

func (l *inviteLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *inviteLogic {
	return &inviteLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *inviteLogic) Path() string {
	return "/gbs/invite/:deviceUniqueId/:channelID/:type" // type: play/playback
}

func (l *inviteLogic) DO() *types.HttpResponse {
	var (
		downloadSpeedQuery         = l.c.Query("downloadSpeed")
		downloadSpeed      float64 = 0
	)
	if downloadSpeedQuery != "" {
		v, err := strconv.ParseFloat(downloadSpeedQuery, 64)
		if err != nil {
			return &types.HttpResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Str("downloadSpeed 参数错误"), localization.M0001),
			}
		}

		downloadSpeed = v
	}

	return l.Invite(&InviteParams{
		DeviceUniqueId: l.c.Param("deviceUniqueId"),
		ChannelID:      l.c.Param("channelID"),
		PlayType:       stream.New().ToPlayType(l.c.Param("type")),
		StartAt:        l.c.Query("startAt"),
		EndAt:          l.c.Query("endAt"),
		Download:       l.c.Query("download") == "1",
		Speed:          downloadSpeed,
		KeepaliveCron:  l.c.Query("keepalive") == "1",
		StreamName:     l.c.Query("streamName"),
		Caller:         "http 请求 gbs/invite",
	})
}

func (l *inviteLogic) Invite(args *InviteParams) *types.HttpResponse {
	var streamName = stream.New().Produce(args.DeviceUniqueId, args.ChannelItem.UniqueId, args.PlayType)
	if args.StreamName != "" {
		streamName = args.StreamName
	}

	// 请求控制 防止重复发送 信令
	l.svcCtx.InviteRequestLock.Lock()
	if l.svcCtx.InviteRequestState.Contains(streamName) {
		l.svcCtx.InviteRequestLock.Unlock()
		return nil
	}

	l.svcCtx.InviteRequestState.Add(streamName)
	l.svcCtx.InviteRequestLock.Unlock()
	defer l.svcCtx.InviteRequestState.Remove(streamName)

	if args.DeviceUniqueId == "" {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("device 参数错误"), localization.M0001),
		}
	}

	if args.ChannelID == "" {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("channel 参数错误"), localization.M0001),
		}
	}

	if !stream.New().PlayTypeVerify(args.PlayType) {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("type 参数错误"), localization.M0001),
		}
	}

	sipReqRes, ok := l.svcCtx.SipCatalogLoopMap.Get(args.DeviceUniqueId)
	if !ok {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(constants.DeviceUnregistered), localization.M00300),
		}
	}

	if args.ChannelItem == nil {
		// 获取设备信息
		channelRes, err := response.NewRpcToHttpResp[*backendservice.Response, *channels.Item]().Parse(
			func() (*backendservice.Response, error) {
				params, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
					Conditions: []*orm.ConditionItem{
						{Column: channels.ColumnUniqueId, Value: args.ChannelID},
						{Column: channels.ColumnDeviceUniqueId, Value: args.DeviceUniqueId},
					},
				})
				if err != nil {
					return nil, response.NewMakeRpcRetErr(err, 2)
				}

				return l.svcCtx.RpcClients.Device.ChannelRowFind(l.ctx, params)
			},
		)
		if err != nil {
			return &types.HttpResponse{Err: err}
		}

		args.ChannelItem = channelRes.Data
	}

	if args.DeviceItem == nil {
		// 获取设备信息
		deviceRes, err := response.NewRpcToHttpResp[*backendservice.Response, *devices.Item]().Parse(
			func() (*backendservice.Response, error) {
				data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
					Conditions: []*orm.ConditionItem{
						{Column: devices.ColumnDeviceUniqueId, Value: args.DeviceUniqueId},
					},
				})
				if err != nil {
					return nil, err
				}

				return l.svcCtx.RpcClients.Device.DeviceRow(l.ctx, data)
			},
		)
		if err != nil {
			return &types.HttpResponse{Err: err}
		}

		args.DeviceItem = deviceRes.Data
	}

	streamRes, err2 := ms.New(l.ctx, l.svcCtx).StreamInfo(args.DeviceItem)
	if err2 != nil {
		return &types.HttpResponse{Err: response.MakeError(response.NewHttpRespMessage().Err(err2), localization.MR1008)}
	}

	// 检测流状态
	groupInDetailResp, _, _ := ms.New(l.ctx, l.svcCtx).GetStreamGroup(streamRes.MediaServerUrl, streamName)
	if groupInDetailResp != nil && groupInDetailResp.Pub != nil && groupInDetailResp.Pub.SessionID != "" {
		return nil
	}

	if l.svcCtx.Config.Sip.MediaServerStreamPortMax <= 0 || l.svcCtx.Config.Sip.MediaServerStreamPortMin <= 0 {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("推流端口范围未设置"), localization.M0001),
		}
	}

	// 流已存在 防止流占用
	if l.svcCtx.PubStreamExistsState.Contains(streamName) && (groupInDetailResp == nil || groupInDetailResp.Pub == nil || groupInDetailResp.Pub.SessionID == "") {
		// 检测输入型 pub 是否存在 不存在调用stream_stop(检测超时)
		var msId uint64 = 0
		if len(args.DeviceItem.MSIds) > 0 {
			msId = args.DeviceItem.MSIds[0]
		}

		// 停止流
		if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode}).HttpPostJsonResJson(
			fmt.Sprintf("http://127.0.0.1:%d/api/video/stream/stop", l.svcCtx.Config.Http.Port),
			map[string]interface{}{
				"streamNames": streamName,
				"msId":        msId,
			},
			nil,
		); err != nil {
			return &types.HttpResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001),
			}
		}
	}

	var msIP = streamRes.MSNode.InternalIP
	if l.svcCtx.Config.Sip.UseExternalWan {
		msIP = streamRes.MSNode.ExtIP
	}

	// 请求ms分配拉流信息
	l.svcCtx.SipSendVideoLiveInvite <- &types.SipVideoLiveInviteMessage{
		StreamPort:     streamRes.StreamPort,
		MediaTransMode: streamRes.TransportProtocol.MediaTransMode,
		MediaServerUrl: streamRes.MediaServerUrl,

		MediaServerIP:   msIP,
		MediaServerPort: streamRes.MSNode.HttpPort,

		StreamName:        streamName,
		PlayType:          args.PlayType,
		DeviceUniqueId:    args.DeviceUniqueId,
		MediaProtocolMode: streamRes.TransportProtocol.MediaProtocolMode,
		ChannelUniqueId:   args.ChannelItem.UniqueId,
		StartAt:           args.StartAt,
		EndAt:             args.EndAt,
		Req:               sipReqRes.Req,
		TransportProtocol: args.DeviceItem.TransportProtocol(),
		Download:          args.Download,
		Speed:             args.Speed,
		Caller:            args.Caller,

		Data: args.DeviceItem,
	}

	return nil
}
