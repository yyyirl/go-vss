package base

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
)

var (
	_ types.HttpEHandleLogic[*statusLogic] = (*statusLogic)(nil)

	StatusLogic = new(statusLogic)
)

type statusLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *statusLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *statusLogic {
	return &statusLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *statusLogic) DO() *types.HttpResponse {
	return &types.HttpResponse{
		Data: map[string]interface{}{
			"bind-host":             l.svcCtx.Config.Host,
			"http-port":             l.svcCtx.Config.Http.Port,
			"sip-port":              l.svcCtx.Config.Sip.Port,
			"sip-password":          l.svcCtx.Config.Sip.Password,
			"sip-use-password":      l.svcCtx.Config.Sip.UsePassword,
			"sip-catalog-interval":  l.svcCtx.Config.Sip.CatalogInterval,
			"sip-heartbeat-timeout": l.svcCtx.Config.Sip.HeartbeatTimeout,
		},
	}
}

func (l *statusLogic) Path() string {
	return "/status"
}
