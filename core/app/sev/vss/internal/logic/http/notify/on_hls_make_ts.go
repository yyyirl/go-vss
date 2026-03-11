// @Title        RTMP停止推流通知
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package notify

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
)

var (
	_ types.HttpRHandleLogic[*OnHlsMakeTsLogic, types.NotifyStreamReq] = (*OnHlsMakeTsLogic)(nil)

	VOnHlsMakeTsLogic = new(OnHlsMakeTsLogic)
)

type OnHlsMakeTsLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnHlsMakeTsLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnHlsMakeTsLogic {
	return &OnHlsMakeTsLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnHlsMakeTsLogic) Path() string {
	return "/notify/on-hls-make-ts"
}

func (l *OnHlsMakeTsLogic) DO(_ types.NotifyStreamReq) *types.HttpResponse {
	return nil
}
