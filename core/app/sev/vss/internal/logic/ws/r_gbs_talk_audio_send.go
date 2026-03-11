package ws

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/audio"
	"skeyevss/core/tps"
)

const GbsTalkAudioSendKey = "gbs-talk-audio-send"

type RGBSTalkAudioSendLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	client *types.WSClient
}

func NewRGBSTalkAudioSend(ctx context.Context, svcCtx *types.ServiceContext, client *types.WSClient) *RGBSTalkAudioSendLogic {
	return &RGBSTalkAudioSendLogic{ctx: ctx, svcCtx: svcCtx, client: client}
}

func (l *RGBSTalkAudioSendLogic) Do(req *types.WSGBSTalkAudioSendReq) *types.WSResponse {
	var (
		key = req.DeviceUniqueId
		// 清理占用状态
		clearUsageStatus = func() {
			l.svcCtx.WSTalkUsageStatus.Remove(key)
			// 通知其他客户端占用状态已被解除
			BGBSSendTalkUsageStatus(l.svcCtx, req.UniqueId, key, 0)
		}
	)
	// 占用状态检测
	v, ok := l.svcCtx.WSTalkUsageStatus.Get(key)
	if !ok {
		l.svcCtx.WSTalkUsageStatus.Set(key, req.UniqueId)
	} else {
		if v != req.UniqueId {
			clearUsageStatus()
			return &types.WSResponse{
				Errors: &tps.XError{Message: "通道对讲已被占用, 等待结束后使用"},
			}
		}
	}

	// 通知其他客户端语音已被占用
	BGBSSendTalkUsageStatus(l.svcCtx, req.UniqueId, key, 1)
	defer clearUsageStatus()

	// 状态检测
	var wsTalkSipStatus *audio.TalkSessionItem
	if v, ok := l.svcCtx.TalkSipData.Get(key); ok {
		wsTalkSipStatus = v
	} else {
		// 通知当前客户端重新获取状态
		go func() {
			time.Sleep(time.Second)
			l.svcCtx.WSProc.ResponseMessageChan <- &types.WSResponseMessageItem{
				Client: l.client,
				Content: &types.WSResponseMessage{
					MessageType: websocket.TextMessage,
					WSResponse: &types.WSResponse{
						Type: GbsTalkSipPubStateKey,
						Data: makeRGBSTalkSipPubState(0),
					},
				},
			}
		}()

		return &types.WSResponse{
			Errors: &tps.XError{Message: "sip状态没有准备成功"},
		}
	}

	// 音频转换
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(req.Stream))
	if err != nil {
		return &types.WSResponse{
			Errors: &tps.XError{Message: "音频文件转换失败"},
		}
	}

	// 更新sip活跃
	if wsTalkSipStatus != nil {
		wsTalkSipStatus.ActivateAt = time.Now().UnixMilli()
		l.svcCtx.TalkSipData.Set(key, wsTalkSipStatus)
	}

	// 发送语音消息
	l.svcCtx.SipSendTalk <- &types.GBSSipSendTalk{
		DeviceUniqueId: req.DeviceUniqueId,
		Data:           data,
	}
	return nil
}
