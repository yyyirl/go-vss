// @Title        discover
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package onvif

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	onvifdevice "github.com/use-go/onvif/device"
	sdkdevice "github.com/use-go/onvif/sdk/device"

	"skeyevss/core/app/sev/vss/internal/pkg/onvif"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

var (
	_ types.HttpRHandleLogic[*DeviceInfoLogic, types.OnvifDeviceInfoReq] = (*DeviceInfoLogic)(nil)

	VDeviceInfoLogic = new(DeviceInfoLogic)
)

type DeviceInfoLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *DeviceInfoLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *DeviceInfoLogic {
	return &DeviceInfoLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *DeviceInfoLogic) Path() string {
	return "/onvif/device-info"
}

// 获取设备信息

func (l *DeviceInfoLogic) DO(req types.OnvifDeviceInfoReq) *types.HttpResponse {
	dev, err := onvif.GetDev(
		&onvif.GetDevParams{
			Username: req.Username,
			Password: req.Password,
			IP:       req.IP,
			Port:     req.Port,
		},
	)
	if err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("Failed to get onvif[1] device info: %v", err)), localization.MR1003),
		}
	}

	deviceInfo, err := sdkdevice.Call_GetDeviceInformation(l.ctx, dev, onvifdevice.GetDeviceInformation{})
	if err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("Failed to get onvif Call_GetDeviceInformation device info: %v", err)), localization.MR1003),
		}
	}

	return &types.HttpResponse{
		Data: deviceInfo,
	}
}
