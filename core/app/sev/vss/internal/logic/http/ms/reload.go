package ms

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/pkg/ms"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

var (
	_ types.HttpRHandleLogic[*ReloadLogic, types.MsReloadReq] = (*ReloadLogic)(nil)

	VReloadLogic = new(ReloadLogic)
)

type ReloadLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *ReloadLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *ReloadLogic {
	return &ReloadLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *ReloadLogic) Path() string {
	return "/ms/reload"
}

func (l *ReloadLogic) DO(req types.MsReloadReq) *types.HttpResponse {
	if err := ms.New(l.ctx, l.svcCtx).Reload(fmt.Sprintf("http://%s:%d", req.IP, req.Port), req); err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010),
		}
	}

	return nil
}
