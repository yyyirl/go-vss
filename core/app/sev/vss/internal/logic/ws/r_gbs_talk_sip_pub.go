package ws

import (
	"context"
	"time"

	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/audio"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/devices"
	"skeyevss/core/repositories/models/dictionaries"
	"skeyevss/core/tps"
)

const GbsTalkSipPubKey = "gbs-talk-sip-pub"

type RGBSTalkSipPubLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	client *types.WSClient
}

func NewRGBSTalkSipPub(ctx context.Context, svcCtx *types.ServiceContext, client *types.WSClient) *RGBSTalkSipPubLogic {
	return &RGBSTalkSipPubLogic{ctx: ctx, svcCtx: svcCtx, client: client}
}

func (l *RGBSTalkSipPubLogic) Do(req *types.WSGBSTalkSipPub) *types.WSResponse {
	var key = req.DeviceUniqueId
	if l.svcCtx.TalkSipSendStatus.Contains(key) {
		// 有其他链接正在发送sip 消息
		BGBSSendTalkPub(l.svcCtx, key, 1)
		return nil
	}

	// 停止语音消息
	l.svcCtx.SipSendTalk <- &types.GBSSipSendTalk{
		DeviceUniqueId: req.DeviceUniqueId,
		Stop:           true,
		StopCaller:     functions.Caller(2),
	}
	time.Sleep(300 * time.Millisecond)
	BGBSSendTalkPub(l.svcCtx, key, 2)

	// 发送sip状态锁
	l.svcCtx.TalkSipSendStatus.Add(key)
	defer l.svcCtx.TalkSipSendStatus.Remove(key)

	// 检测设备在线状态
	res, ok := l.svcCtx.SipCatalogLoopMap.Get(req.DeviceUniqueId)
	if !ok {
		return &types.WSResponse{
			Errors: &tps.XError{Message: "设备不在线"},
		}
	}

	// 记录sip状态
	l.svcCtx.TalkSipData.Set(key, &audio.TalkSessionItem{
		Status:     false,
		ActivateAt: time.Now().UnixMilli(),
	})

	// 获取设备消息
	deviceRowRes, err1 := response.NewRpcToHttpResp[*deviceservice.Response, *devices.Item]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: devices.ColumnDeviceUniqueId, Value: req.DeviceUniqueId},
				},
			})
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.DeviceRow(l.ctx, data)
		},
	)
	if err1 != nil {
		return &types.WSResponse{
			Errors: &tps.XError{Message: err1.Error},
		}
	}

	// 大华设备
	var dahuaState = false
	for _, item := range l.svcCtx.DictionaryMap[dictionaries.UniqueIdDeviceManufacturer].Children {
		if item.Raw == nil {
			continue
		}

		if deviceRowRes.Data.ManufacturerId == item.Raw.ID && functions.Contains(
			item.Raw.UniqueId,
			[]string{dictionaries.UniqueIdDeviceManufacturer_3, dictionaries.UniqueIdDeviceManufacturer_7},
		) {
			dahuaState = true
			break
		}
	}

	if dahuaState {
		// 发送invite
		l.svcCtx.SipSendTalkInvite <- &types.SipTalkInviteMessage{
			ChannelUniqueId: req.ChannelUniqueId,
			DeviceUniqueId:  req.DeviceUniqueId,
			Req:             res.Req,
			DeviceRow:       deviceRowRes.Data,
		}
	} else {
		// 发送广播数据
		l.svcCtx.SipSendBroadcast <- &types.BroadcastReq{
			ChannelUniqueId: req.ChannelUniqueId,
			DeviceUniqueId:  req.DeviceUniqueId,
			Req:             res.Req,
		}
	}

	return nil
}
