package ws

import (
	"skeyevss/core/app/sev/vss/internal/logic/ws"
	"skeyevss/core/app/sev/vss/internal/types"
)

type (
	routeHandleType     func(svcCtx *types.ServiceContext, data *types.WSHandlerCallParams) *types.WSResponse
	broadcastHandleType func(svcCtx *types.ServiceContext, data interface{}) error
)

type (
	routeItemType struct {
		handler routeHandleType
	}

	broadcastItemType struct {
		handler broadcastHandleType
	}
)

var (
	// 请求路由 type
	routers = map[string]*routeItemType{
		"heartbeat": {
			handler: func(_ *types.ServiceContext, _ *types.WSHandlerCallParams) *types.WSResponse {
				return nil
			},
		},
		// 发送对讲语音
		ws.GbsTalkAudioSendKey: {
			handler: func(svcCtx *types.ServiceContext, params *types.WSHandlerCallParams) *types.WSResponse {
				var data types.WSGBSTalkAudioSendReq
				if err := params.RequestParse(params.Req, &data); err != nil {
					return err
				}

				return ws.NewRGBSTalkAudioSend(params.Ctx, svcCtx, params.Client).Do(&data)
			},
		},
		// 停止对讲语音
		// 注册sip -> Broadcast -> Invite -> ACK
		// 注册sip 状态
		// 注册语音通道接收广播
	}

	// 全局广播
	broadcasters = map[string]*broadcastItemType{
		// 语音对讲sip状态通知
		// 语音使用状态
	}
)

// 1 创建链接
// 2 注册对讲 方便通知ack状态
// 3 发送invite
// 4 开始对讲
// 5 20s没有操作结束invite
// 6 主动停止gbs-talk-stop 关闭网页
