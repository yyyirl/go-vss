// @Title        流媒体数据更新 编码格式/音视频参数等
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package notify

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
)

var (
	_ types.HttpRHandleLogic[*OnUpdateLogic, types.NotifyStreamReq] = (*OnUpdateLogic)(nil)

	VOnUpdateLogic = new(OnUpdateLogic)
)

type OnUpdateLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnUpdateLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnUpdateLogic {
	return &OnUpdateLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnUpdateLogic) Path() string {
	return "/notify/on-update"
}

func (l *OnUpdateLogic) DO(_ types.NotifyStreamReq) *types.HttpResponse {
	return nil
}
