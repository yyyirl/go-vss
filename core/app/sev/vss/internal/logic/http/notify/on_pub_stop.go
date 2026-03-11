// @Title        RTMP停止推流通知(当前做为上级,下级(设备)给当前推流)
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package notify

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
)

var (
	_ types.HttpRHandleLogic[*OnPubStopLogic, types.NotifyStreamReq] = (*OnPubStopLogic)(nil)

	VOnPubStopLogic = new(OnPubStopLogic)
)

type OnPubStopLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnPubStopLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnPubStopLogic {
	return &OnPubStopLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnPubStopLogic) Path() string {
	return "/notify/on-pub-stop"
}

func (l *OnPubStopLogic) DO(req types.NotifyStreamReq) *types.HttpResponse {
	l.svcCtx.PubStreamExistsState.Remove(req.StreamName)
	return setStreamState(l.ctx, l.c, l.svcCtx, req, 0, l.Path())
}
