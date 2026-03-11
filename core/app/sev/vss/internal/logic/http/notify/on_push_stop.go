// @Title        RTMP停止推流通知(当前作为下级(设备),给上级推流)
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package notify

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/common/stream"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

var (
	_ types.HttpRHandleLogic[*OnPushStopLogic, types.NotifyStreamReq] = (*OnPushStopLogic)(nil)

	VOnPushStopLogic = new(OnPushStopLogic)
)

type OnPushStopLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *OnPushStopLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *OnPushStopLogic {
	return &OnPushStopLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *OnPushStopLogic) Path() string {
	return "/notify/on-push-stop"
}

func (l *OnPushStopLogic) DO(req types.NotifyStreamReq) *types.HttpResponse {
	data, err := stream.New().Parse(req.StreamName)
	if err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.MR1004),
		}
	}

	if data.Channel == "" || data.Device == "" {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.MR1004),
		}
	}

	for _, value := range l.svcCtx.GBCInviteReqMaps.All() {
		if value.SessionId == req.SessionId && value.StreamName == req.StreamName {
			// TODO 完整版请联系作者
			// GBC通知
			break
		}
	}

	return nil
}
