package ws

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"skeyevss/core/app/sev/vss/internal/logic/ws"
	"skeyevss/core/app/sev/vss/internal/pkg"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/tps"
)

type Interval struct {
	svcCtx *types.ServiceContext

	fetchLock sync.Mutex
}

func NewInterval(svcCtx *types.ServiceContext) *Interval {
	return &Interval{svcCtx: svcCtx}
}

func (l *Interval) Do() {
	go l.heartbeat()
	go l.tokenVerify()
	go l.clearTalkSipStatus()
}

// 心跳
func (l *Interval) heartbeat() {
	defer pkg.NewRecover(&l.svcCtx.Config, l.heartbeat)

	for v := range time.NewTicker(time.Second * 1).C {
		l.svcCtx.WSClientCache.Range(func(client *types.WSClient) {
			if client == nil {
				return
			}

			if client.IsClosed {
				NewProc(l.svcCtx).closer(&types.WSCloseChanItem{
					Error:  tps.NewErr("链接已关闭"),
					Client: client,
				})
				return
			}

			// 检测链接活跃度 客户端需要每{ MaxLifetime }发送一次心跳
			if int64(math.Abs(float64(v.Unix()-client.ActiveTime))) >= l.svcCtx.Config.WS.MaxLifetime {
				NewProc(l.svcCtx).closer(&types.WSCloseChanItem{
					Error:  tps.NewErr(fmt.Sprintf("断开连接 链接不再活跃 client id: %v", client.ClientId)),
					Client: client,
				})
				return
			}

			// 检验合法性 从创建链接开始计时 超过最大鉴权时间未登录则断开连接
			if int64(math.Abs(float64(v.Unix()-client.ConnTime))) >= l.svcCtx.Config.WS.AuthorizationLifetime && !client.Validate {
				NewProc(l.svcCtx).closer(&types.WSCloseChanItem{
					Error:  tps.NewErr(fmt.Sprintf("断开连接 链接长时间未登录 client id: %v", client.ClientId)),
					Client: client,
				})
				return
			}
		})
	}
}

// token校验
func (l *Interval) tokenVerify() {
	defer pkg.NewRecover(&l.svcCtx.Config, l.tokenVerify)

	for range time.NewTicker(time.Second * 20).C {
		l.svcCtx.WSClientCache.Range(func(client *types.WSClient) {
			if client == nil {
				return
			}

			if client.IsClosed {
				NewProc(l.svcCtx).closer(&types.WSCloseChanItem{
					Error:  tps.NewErr("链接已关闭"),
					Client: client,
				})
				return
			}

			if client.Token == "" {
				return
			}

			// 长时间未刷新token
			if _, err := pkg.NewAes(l.svcCtx.Config).ParseUserToken(client.Token); err != nil {
				// 发送登录超时消息
				l.svcCtx.WSProc.ResponseMessageChan <- &types.WSResponseMessageItem{
					Client: client,
					Content: &types.WSResponseMessage{
						MessageType: websocket.TextMessage,
						WSResponse: &types.WSResponse{
							Type:   "login",
							Errors: tps.NewErr("登录超时"),
						},
					},
					AlterCall: func() {
						// 关闭连接
						l.svcCtx.WSProc.CloseChan <- &types.WSCloseChanItem{
							Client: client,
							Error:  tps.NewErr("登录超时连接已被关闭"),
						}
					},
				}
			}
		})
	}
}

// 清理对讲sip状态 超过ClearTalkSipInterval秒不活跃
func (l *Interval) clearTalkSipStatus() {
	defer pkg.NewRecover(&l.svcCtx.Config, l.tokenVerify)

	for v := range time.NewTicker(time.Second * 1).C {
		var (
			now  = v.UnixMilli()
			maps = l.svcCtx.TalkSipData.All()
		)
		for key, item := range maps {
			if item.ActivateAt > 0 && now-item.ActivateAt >= l.svcCtx.Config.WS.ClearTalkSipInterval {
				ws.RGBSTalkAudioStop(l.svcCtx, key)
			}
		}
	}
}
