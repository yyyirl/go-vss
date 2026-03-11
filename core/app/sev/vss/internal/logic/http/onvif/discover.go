// @Title        discover
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package onvif

import (
	"context"
	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/types"
)

var (
	_ types.HttpEHandleLogic[*discoverStreamLogic] = (*discoverStreamLogic)(nil)

	DiscoverLogic = new(discoverStreamLogic)
)

type discoverStreamLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *discoverStreamLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *discoverStreamLogic {
	return &discoverStreamLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *discoverStreamLogic) Path() string {
	return "/onvif/discover"
}

// 发现局域网内的ONVIF设备

func (l *discoverStreamLogic) DO() *types.HttpResponse {
	return &types.HttpResponse{Data: l.svcCtx.OnvifDiscoverDevices}
}
