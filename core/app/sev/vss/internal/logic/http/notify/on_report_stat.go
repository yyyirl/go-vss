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
	_ types.HttpRHandleLogic[*OnReportStatLogic, types.NotifyStreamReq] = (*OnReportStatLogic)(nil)

	VOnReportStatLogic = new(OnReportStatLogic)
)

type OnReportStatLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnReportStatLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnReportStatLogic {
	return &OnReportStatLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnReportStatLogic) Path() string {
	return "/notify/on-report-stat"
}

func (l *OnReportStatLogic) DO(_ types.NotifyStreamReq) *types.HttpResponse {
	return nil
}
