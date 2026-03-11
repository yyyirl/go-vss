// @Title        生成wstoken
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package base

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/config"
	"skeyevss/core/app/sev/vss/internal/pkg"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

var (
	_ types.HttpRHandleLogic[*WSTokenLogic, types.WSTokenReq] = (*WSTokenLogic)(nil)

	VWSTokenLogic = new(WSTokenLogic)
)

type WSTokenLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *WSTokenLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *WSTokenLogic {
	return &WSTokenLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *WSTokenLogic) Path() string {
	return "/ws-token"
}

func (l *WSTokenLogic) DO(req types.WSTokenReq) *types.HttpResponse {
	authorization, err := pkg.NewAes(config.Config{XAuth: l.svcCtx.Config.XAuth}).MakeXAuthorization(req.ID)
	if err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("token 生成失败"), localization.M0002),
		}
	}

	return &types.HttpResponse{
		Data: authorization,
	}
}
