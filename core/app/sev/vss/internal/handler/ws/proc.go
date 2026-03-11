package ws

import (
	"fmt"

	"github.com/gorilla/websocket"

	"skeyevss/core/app/sev/vss/internal/pkg"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/tps"
)

type Proc struct {
	svcCtx *types.ServiceContext
}

func NewProc(svcCtx *types.ServiceContext) *Proc {
	return &Proc{
		svcCtx: svcCtx,
	}
}

func (p *Proc) reader(client *types.WSClient) {
	defer pkg.NewRecover(&p.svcCtx.Config, func() {
		p.reader(client)
	})

	for {
		msgType, data, err := client.WebsocketConn.ReadMessage()
		switch {
		case websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived):
			functions.LogInfo("client", client.ClientId, " read loop IsCloseError 关闭消息读取")
			if client.IsClosed {
				return
			}

			// 发送关闭消息
			p.svcCtx.WSProc.CloseChan <- &types.WSCloseChanItem{
				Error:  tps.NewErr("websocket close error: " + err.Error()),
				Client: client,
			}
			return

		case websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure):
			functions.LogInfo("client", client.ClientId, " read loop IsUnexpectedCloseError 关闭消息读取")
			if client.IsClosed {
				return
			}

			p.svcCtx.WSProc.CloseChan <- &types.WSCloseChanItem{
				Error:  tps.NewErr("websocket unexpected close error: " + err.Error()),
				Client: client,
			}
			return

		case msgType == websocket.TextMessage: // 消息读取成功
			if client.IsClosed {
				return
			}

			// 设置活跃时间
			client.ActiveTime = functions.NewTimer().Now()
			// 发送到处理消息队列中
			p.svcCtx.WSProc.ReceiveMessageChan <- &types.WSMessageReceiveItem{
				Client: client,
				Content: &types.WSReceiveMessage{
					MessageType: msgType,
					Content:     data,
				},
			}

		default:
			if client.IsClosed {
				return
			}

			// if err != nil && strings.Index(err.Error(), "unexpected reserved bits 0x") >= 0 {
			// 	//  websocket: unexpected reserved bits 0x40
			// 	// println("消息读取失败 err: " + err.Error())
			// 	continue
			// }

			// 关闭链接
			p.svcCtx.WSProc.CloseChan <- &types.WSCloseChanItem{
				Error:  tps.NewErr(fmt.Sprintf("消息读取失败, 未知类型 关闭链接, userid: %v", client.Userid)),
				Client: client,
			}
			return
		}
	}
}

func (p *Proc) Receiver() {
	defer pkg.NewRecover(&p.svcCtx.Config, p.Receiver)

	for {
		select {
		case data := <-p.svcCtx.WSProc.ReceiveMessageChan: // 接收消息
			if data == nil || data.Client == nil {
				continue
			}

			if data.Client.IsClosed {
				continue
			}

			if resp := p.dispatcher(data); resp != nil && resp.WSResponse != nil {
				p.svcCtx.WSProc.ResponseMessageChan <- &types.WSResponseMessageItem{
					Client:  data.Client,
					Content: resp,
				}
			}

		case data := <-p.svcCtx.WSProc.ResponseMessageChan: // 响应消息
			if data == nil || data.Client == nil {
				continue
			}

			if data.Client.IsClosed {
				continue
			}

			if data.Content == nil {
				continue
			}

			if err := newResponse(p.svcCtx, data.Client).current(data.Content); err != nil {
				if data.Client.IsClosed {
					continue
				}

				p.svcCtx.WSProc.CloseChan <- &types.WSCloseChanItem{
					Client: data.Client,
					Error:  tps.NewErr(err.Error()),
				}

				continue
			}

			if data.AlterCall != nil {
				data.AlterCall()
			}

		case data := <-p.svcCtx.WSProc.BroadcastChan: // 全局广播
			if err := newBroadcaster(p.svcCtx).dispatch(data); err != nil {
				functions.LogError("broadcast", "广播消息执行失败, err:", err.Error())
			}

		case data := <-p.svcCtx.WSProc.CloseChan: // 关闭链接
			if data == nil || data.Client == nil {
				continue
			}

			p.closer(data)
			continue
		}
	}
}

// 关闭链接
func (p *Proc) closer(data *types.WSCloseChanItem) {
	// 设置关闭
	data.Client.IsClosed = true
	// close
	data.Client.CloseChanSignal.Do(func() {
		// 删除缓存
		p.svcCtx.WSClientCache.Delete(data.Client)
		// 关闭链接
		_ = data.Client.WebsocketConn.Close()
		functions.LogInfo("链接已被关闭 userid: ", data.Client.Userid, ", err: ", data.Error.Error(), ", close caller: ", functions.Caller(5))
	})
}

func (p *Proc) dispatcher(data *types.WSMessageReceiveItem) *types.WSResponseMessage {
	request, err := p.parser(data.Content)
	if err != nil {
		return &types.WSResponseMessage{
			MessageType: data.Content.MessageType,
			WSResponse: &types.WSResponse{
				Errors: tps.NewErr(err.Error()),
			},
		}
	}

	return newRouters(p.svcCtx, data.Client).dispatch(request)
}

// 解析接收到的消息
func (p *Proc) parser(req *types.WSReceiveMessage) (*types.WSRequestContent, error) {
	var request types.WSRequestContent
	if err := functions.JSONUnmarshal(req.Content, &request); err != nil {
		return nil, err
	}

	request.MessageType = req.MessageType
	return &request, nil
}
