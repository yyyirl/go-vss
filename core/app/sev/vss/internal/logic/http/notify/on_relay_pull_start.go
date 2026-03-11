// @Title        开始拉流通知
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package notify

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
)

var (
	_ types.HttpRHandleLogic[*OnRelayPullStartLogic, types.NotifyStreamReq] = (*OnRelayPullStartLogic)(nil)

	VOnRelayPullStartLogic = new(OnRelayPullStartLogic)
)

type OnRelayPullStartLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnRelayPullStartLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnRelayPullStartLogic {
	return &OnRelayPullStartLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnRelayPullStartLogic) Path() string {
	return "/notify/on-reply-pull-start"
}

func (l *OnRelayPullStartLogic) DO(req types.NotifyStreamReq) *types.HttpResponse {
	return setStreamState(l.ctx, l.c, l.svcCtx, req, 1, l.Path())
}
