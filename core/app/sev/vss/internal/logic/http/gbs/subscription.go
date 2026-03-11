package gbs

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

var (
	_ types.HttpRHandleLogic[*SubscriptionLogic, types.SubscriptionReq] = (*SubscriptionLogic)(nil)

	VSubscriptionLogic = new(SubscriptionLogic)
)

type SubscriptionLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *SubscriptionLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *SubscriptionLogic {
	return &SubscriptionLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *SubscriptionLogic) Path() string {
	return "/gbs/subscription"
}

func (l *SubscriptionLogic) DO(req types.SubscriptionReq) *types.HttpResponse {
	if req.DeviceUniqueId == "" {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("参数错误"), localization.M0001),
		}
	}

	l.svcCtx.SipSendSubscription <- &req

	return nil
}
