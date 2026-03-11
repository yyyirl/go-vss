package ws

import (
	"context"
	"fmt"
	"time"

	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
)

const GbsTalkAudioStopKey = "gbs-talk-audio-stop"

type RGBSTalkAudioStopLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	client *types.WSClient
}

func NewRGBSTalkAudioStop(ctx context.Context, svcCtx *types.ServiceContext, client *types.WSClient) *RGBSTalkAudioStopLogic {
	return &RGBSTalkAudioStopLogic{ctx: ctx, svcCtx: svcCtx, client: client}
}

func (l *RGBSTalkAudioStopLogic) Do(req *types.WSGBSTalkAudioStopReq) *types.WSResponse {
	RGBSTalkAudioStop(l.svcCtx, fmt.Sprintf("%s-%s", req.DeviceUniqueId, req.ChannelUniqueId))
	return nil
}

func RGBSTalkAudioStop(svcCtx *types.ServiceContext, key string) {
	// 停止语音消息
	svcCtx.SipSendTalk <- &types.GBSSipSendTalk{
		DeviceUniqueId: key,
		Stop:           true,
		StopCaller:     functions.Caller(2),
	}

	time.Sleep(300 * time.Millisecond)
	// 广播重置消息
	BGBSSendTalkPub(svcCtx, key, 0)
	// 广播占用状态已被解除
	BGBSSendTalkUsageStatus(svcCtx, "", key, 0)
	// 清理状态
	svcCtx.CloseWSTalkSip(key)
}
