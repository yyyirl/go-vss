// @Title        i帧
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package notify

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
)

var (
	_ types.HttpRHandleLogic[*OnReportFrameInfoLogic, types.NotifyStreamReq] = (*OnReportFrameInfoLogic)(nil)

	VOnReportFrameInfoLogic = new(OnReportFrameInfoLogic)
)

type OnReportFrameInfoLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnReportFrameInfoLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnReportFrameInfoLogic {
	return &OnReportFrameInfoLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnReportFrameInfoLogic) Path() string {
	return "/notify/report-frame-info"
}

func (l *OnReportFrameInfoLogic) DO(_ types.NotifyStreamReq) *types.HttpResponse {
	return nil
}
