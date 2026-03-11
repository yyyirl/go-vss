// @Title        sip日志
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package sse

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

var (
	_ types.SSEHandleSPLogic[*SipLogLogic] = (*SipLogLogic)(nil)

	SipLogsType = "sip_logs"

	VSipLogs = new(SipLogLogic)

	sipLogIsActive atomic.Bool
)

type SipLogLogic struct {
	ctx         context.Context
	svcCtx      *types.ServiceContext
	messageChan chan *types.SSEResponse
	close       bool
}

func (l *SipLogLogic) New(ctx context.Context, svcCtx *types.ServiceContext, messageChan chan *types.SSEResponse) *SipLogLogic {
	return &SipLogLogic{
		ctx:         ctx,
		svcCtx:      svcCtx,
		messageChan: messageChan,
	}
}

func (l *SipLogLogic) GetType() string {
	return SipLogsType
}

func (l *SipLogLogic) DO() {
	if sipLogIsActive.Load() {
		l.messageChan <- &types.SSEResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("其他客户端正在使用"), localization.M00274),
		}
		return
	}

	sipLogIsActive.Store(true)

	go l.do()
	go func() {
		var ticker = time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				l.messageChan <- &types.SSEResponse{
					Data: map[string]interface{}{
						"type":    "",
						"content": "heartbeat",
					},
				}

			case <-l.ctx.Done():
				sipLogIsActive.Store(false)
				l.messageChan <- &types.SSEResponse{
					Done: true,
				}
				l.close = true
				l.svcCtx.Broadcast.UnregisterReceiver(types.BroadcastTypeSipRequest)
				l.svcCtx.Broadcast.UnregisterReceiver(types.BroadcastTypeSipReceive)
				return
			}
		}
	}()
}

func (l *SipLogLogic) do() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		var receiver = l.svcCtx.Broadcast.RegisterReceiver(types.BroadcastTypeSipRequest)
		for data := range receiver {
			if l.close {
				return
			}

			l.messageChan <- &types.SSEResponse{
				Data: map[string]interface{}{
					"type":    types.BroadcastTypeSipRequest,
					"content": data.(string),
				},
			}
		}
	}()

	go func() {
		defer wg.Done()

		var receiver = l.svcCtx.Broadcast.RegisterReceiver(types.BroadcastTypeSipReceive)
		for data := range receiver {
			if l.close {
				return
			}

			l.messageChan <- &types.SSEResponse{
				Data: map[string]interface{}{
					"type":    types.BroadcastTypeSipReceive,
					"content": data.(string),
				},
			}
		}
	}()

	wg.Wait()
}
