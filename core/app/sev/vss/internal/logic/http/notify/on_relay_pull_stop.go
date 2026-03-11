// @Title        停止拉流通知
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package notify

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
)

var (
	_ types.HttpRHandleLogic[*OnRelayPullStopLogic, types.NotifyStreamReq] = (*OnRelayPullStopLogic)(nil)

	VOnRelayPullStopLogic = new(OnRelayPullStopLogic)
)

type OnRelayPullStopLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnRelayPullStopLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnRelayPullStopLogic {
	return &OnRelayPullStopLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnRelayPullStopLogic) Path() string {
	return "/notify/on-reply-pull-stop"
}

func (l *OnRelayPullStopLogic) DO(req types.NotifyStreamReq) *types.HttpResponse {
	return setStreamState(l.ctx, l.c, l.svcCtx, req, 0, l.Path())
}
