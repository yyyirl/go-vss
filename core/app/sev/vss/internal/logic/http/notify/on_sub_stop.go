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
	_ types.HttpRHandleLogic[*OnSubStopLogic, types.NotifyStreamReq] = (*OnSubStopLogic)(nil)

	VOnSubStopLogic = new(OnSubStopLogic)
)

type OnSubStopLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnSubStopLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnSubStopLogic {
	return &OnSubStopLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnSubStopLogic) Path() string {
	return "/notify/on-sub-stop"
}

func (l *OnSubStopLogic) DO(_ types.NotifyStreamReq) *types.HttpResponse {
	return nil
}
