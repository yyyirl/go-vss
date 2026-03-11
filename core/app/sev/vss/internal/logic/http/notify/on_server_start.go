// @Title        media server服务启动
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package notify

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
)

var (
	_ types.HttpRHandleLogic[*OnServerStartLogic, types.NotifyStreamReq] = (*OnServerStartLogic)(nil)

	VOnServerStartLogic = new(OnServerStartLogic)
)

type OnServerStartLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnServerStartLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnServerStartLogic {
	return &OnServerStartLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnServerStartLogic) Path() string {
	return "/notify/on-server-start"
}

func (l *OnServerStartLogic) DO(_ types.NotifyStreamReq) *types.HttpResponse {
	return nil
}
