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
	_ types.HttpRHandleLogic[*GetConfigLogic, types.MsGetConfigReq] = (*GetConfigLogic)(nil)

	VGetConfigLogic = new(GetConfigLogic)
)

type GetConfigLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *GetConfigLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *GetConfigLogic {
	return &GetConfigLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *GetConfigLogic) Path() string {
	return "/ms/config"
}

func (l *GetConfigLogic) DO(req types.MsGetConfigReq) *types.HttpResponse {
	resp, err := ms.New(l.ctx, l.svcCtx).GetMSConf1(fmt.Sprintf("http://%s:%d", req.IP, req.Port))
	if err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010),
		}
	}

	return &types.HttpResponse{Data: resp}
}
