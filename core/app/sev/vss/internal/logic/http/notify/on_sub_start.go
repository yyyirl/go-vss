// @Title        RTMP开始推流通知
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package notify

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
)

var (
	_ types.HttpRHandleLogic[*OnSubStartLogic, types.NotifyStreamReq] = (*OnSubStartLogic)(nil)

	VOnSubStartLogic = new(OnSubStartLogic)
)

type OnSubStartLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnSubStartLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnSubStartLogic {
	return &OnSubStartLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnSubStartLogic) Path() string {
	return "/notify/on-sub-start"
}

func (l *OnSubStartLogic) DO(req types.NotifyStreamReq) *types.HttpResponse {
	return setStreamState(l.ctx, l.c, l.svcCtx, req, 1, l.Path())
}
