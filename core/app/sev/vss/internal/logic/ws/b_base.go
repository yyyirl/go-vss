package ws

import (
	"github.com/gorilla/websocket"

	"skeyevss/core/app/sev/vss/internal/types"
)

type broadcasts struct {
	svcCtx *types.ServiceContext
}

func (l *broadcasts) sendWithActivateKey(Type, key string, data interface{}) {
	l.svcCtx.WSClientCache.Range(func(client *types.WSClient) {
		if client == nil {
			return
		}

		if client.IsClosed {
			return
		}

		if client.SipTalkActivateKey != key {
			return
		}

		// 通知所有注册链接使用状态
		l.svcCtx.WSProc.ResponseMessageChan <- &types.WSResponseMessageItem{
			Client: client,
			Content: &types.WSResponseMessage{
				MessageType: websocket.TextMessage,
				WSResponse: &types.WSResponse{
					Type: Type,
					Data: data,
				},
			},
		}
	})
}
