package gbs

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/constants"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

var (
	_ types.HttpEHandleLogic[*catalogLogic] = (*catalogLogic)(nil)

	CatalogLogic = new(catalogLogic)
)

type catalogLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *catalogLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *catalogLogic {
	return &catalogLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *catalogLogic) Path() string {
	return "/gbs/catalog/:deviceUniqueId"
}

func (l *catalogLogic) DO() *types.HttpResponse {
	var deviceUniqueId = l.c.Param("deviceUniqueId")
	if deviceUniqueId == "" {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("参数错误"), localization.M0001),
		}
	}

	res, ok := l.svcCtx.SipCatalogLoopMap.Get(deviceUniqueId)
	if !ok {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(constants.DeviceUnregistered), localization.M00300),
		}
	}

	res.Req.Caller = functions.CallerFile(1)
	l.svcCtx.SipSendCatalog <- res.Req
	return nil
}
