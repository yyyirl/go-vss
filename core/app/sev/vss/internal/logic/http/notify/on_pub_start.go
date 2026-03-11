// @Title        RTMP开始推流通知(当前做为上级,下级(设备)给当前推流)
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package notify

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
)

var (
	_ types.HttpRHandleLogic[*OnPubStartLogic, types.NotifyStreamReq] = (*OnPubStartLogic)(nil)

	VOnPubStartLogic = new(OnPubStartLogic)
)

type OnPubStartLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnPubStartLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnPubStartLogic {
	return &OnPubStartLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnPubStartLogic) Path() string {
	return "/notify/on-pub-start"
}

func (l *OnPubStartLogic) DO(req types.NotifyStreamReq) *types.HttpResponse {
	return setStreamState(l.ctx, l.c, l.svcCtx, req, 1, l.Path())
}
