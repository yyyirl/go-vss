// @Title        有RTMP推流连接建立的事件通知
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package notify

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
)

var (
	_ types.HttpRHandleLogic[*OnRtmpConnectLogic, types.NotifyRtmpConnectReq] = (*OnRtmpConnectLogic)(nil)

	VOnRtmpConnectLogic = new(OnRtmpConnectLogic)
)

type OnRtmpConnectLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnRtmpConnectLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnRtmpConnectLogic {
	return &OnRtmpConnectLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnRtmpConnectLogic) Path() string {
	return "/notify/on-rtmp-connect"
}

func (l *OnRtmpConnectLogic) DO(_ types.NotifyRtmpConnectReq) *types.HttpResponse {
	return nil
}
