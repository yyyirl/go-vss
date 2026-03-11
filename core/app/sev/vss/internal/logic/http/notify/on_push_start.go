// @Title        RTMP开始推流通知(当前作为下级(设备),给上级推流)
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package notify

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
)

var (
	_ types.HttpRHandleLogic[*OnPushStartLogic, types.NotifyStreamReq] = (*OnPushStartLogic)(nil)

	VOnPushStartLogic = new(OnPushStartLogic)
)

type OnPushStartLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnPushStartLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnPushStartLogic {
	return &OnPushStartLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnPushStartLogic) Path() string {
	return "/notify/on-push-start"
}

func (l *OnPushStartLogic) DO(_ types.NotifyStreamReq) *types.HttpResponse {
	return nil
}
