// @Title        discover
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package onvif

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	media "github.com/use-go/onvif/media"
	sdkmedia "github.com/use-go/onvif/sdk/media"

	"skeyevss/core/app/sev/vss/internal/pkg/onvif"
	"skeyevss/core/app/sev/vss/internal/types"
	cTypes "skeyevss/core/common/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

var (
	_ types.HttpRHandleLogic[*DeviceProfilesLogic, types.OnvifDeviceInfoReq] = (*DeviceProfilesLogic)(nil)

	VDeviceProfilesLogic = new(DeviceProfilesLogic)
)

type DeviceProfilesLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *DeviceProfilesLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *DeviceProfilesLogic {
	return &DeviceProfilesLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *DeviceProfilesLogic) Path() string {
	return "/onvif/device-profiles"
}

// 获取设备信息
func (l *DeviceProfilesLogic) DO(req types.OnvifDeviceInfoReq) *types.HttpResponse {
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

	profilesRes, err := sdkmedia.Call_GetProfiles(context.Background(), dev, media.GetProfiles{})
	if err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("Failed to get onvif Call_GetProfiles: %v", err)), localization.MR1003),
		}
	}

	if len(profilesRes.Profiles) == 0 {
		return nil
	}

	var profiles []*cTypes.OnvifDeviceProfileItem
	for _, item := range profilesRes.Profiles {
		var streamUrlReq = media.GetStreamUri{ProfileToken: item.Token}
		streamUrlReq.StreamSetup.Transport.Protocol = "RTSP"
		streamUrlReq.StreamSetup.Stream = "RTP-Unicast"
		streamUrlRes, err := sdkmedia.Call_GetStreamUri(context.Background(), dev, streamUrlReq)
		if err != nil {
			return &types.HttpResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("Failed to get onvif Call_GetStreamUri: %v", err)), localization.MR1003),
			}
		}

		if streamUrlRes.MediaUri.Uri == "" {
			return &types.HttpResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Str("设备通道信息获取失败"), localization.MR1003),
			}
		}

		profiles = append(profiles, &cTypes.OnvifDeviceProfileItem{
			Profile:      string(item.Name),
			ProfileToken: string(item.Token),
			Url:          l.buildPlayUrl(string(streamUrlRes.MediaUri.Uri), req.Username, req.Password),
		})
	}

	return &types.HttpResponse{Data: profiles}
}

func (l *DeviceProfilesLogic) buildPlayUrl(url, username, password string) string {
	if username != "" && password != "" {
		return fmt.Sprintf("rtsp://%s:%s@%s", username, password, strings.TrimLeft(url, "rtsp://"))
	}

	return url
}
