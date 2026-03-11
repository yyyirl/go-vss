package setting

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/common"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/common/source/permissions"
	"skeyevss/core/constants"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/repositories/models/alarms"
	"skeyevss/core/repositories/models/cascade"
	"skeyevss/core/repositories/models/channels"
	"skeyevss/core/repositories/models/devices"
	systemOperationLogs "skeyevss/core/repositories/models/system-operation-logs"
	"skeyevss/core/tps"
)

type SettingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SettingLogic {
	return &SettingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SettingLogic) Setting() interface{} {
	var (
		remoteRes    = l.svcCtx.RemoteReq(l.ctx)
		vssSseUrl    = remoteRes.VssSseUrl
		vssHttpUrl   = remoteRes.VssHttpUrl
		proxyFileUrl = l.svcCtx.Config.SevBase.ProxyFileUrl
		wsUrl        = fmt.Sprintf("ws://%s:%d", l.svcCtx.Config.WebsocketHost, l.svcCtx.Config.WebsocketPort)
	)
	if l.svcCtx.Config.Domain != "" && l.svcCtx.Config.WSProxy != "" {
		if contextx.IsHttps(l.ctx) {
			wsUrl = fmt.Sprintf("wss://%s/%s", l.svcCtx.Config.Domain, l.svcCtx.Config.WSProxy)
		} else {
			wsUrl = fmt.Sprintf("ws://%s/%s", l.svcCtx.Config.Domain, l.svcCtx.Config.WSProxy)
		}
	}

	if l.svcCtx.Config.VssSseTargetFrontend != "" {
		vssSseUrl = l.svcCtx.Config.VssSseTargetFrontend
	}

	if l.svcCtx.Config.VssHttpTargetFrontend != "" {
		vssHttpUrl = l.svcCtx.Config.VssHttpTargetFrontend
	}

	if l.svcCtx.Config.WebProxyFileTargetFrontend != "" {
		proxyFileUrl = l.svcCtx.Config.WebProxyFileTargetFrontend
	}

	var settingContent = l.svcCtx.Settings().Content
	if settingContent != nil {
		if settingContent.MapZoom <= 6 || settingContent.MapZoom > 12 {
			settingContent.MapZoom = 6
		}
	}

	return &common.Setting{
		Showcase:                contextx.GetShowcaseState(l.ctx),
		Super:                   contextx.GetSuperState(l.ctx),
		SystemOperationLogTypes: systemOperationLogs.TypeViews,
		Permissions:             permissions.Source(),
		PermissionIds:           contextx.GetPermissionIds(l.ctx),
		Setting:                 settingContent,
		InternalIP:              l.svcCtx.Config.InternalIP,
		MediaTransModes:         devices.MediaTransModeMaps,
		AccessProtocols:         devices.AccessProtocols,
		AccessProtocolColors:    devices.AccessProtocolColors,
		ChannelFilters:          devices.ChannelFilters,
		BitstreamIndexes:        devices.BitstreamIndexes,
		PTXTypes:                channels.PTXTypes,
		VssHttpUrl:              vssHttpUrl,
		VssSseUrl:               vssSseUrl,
		MSUrl:                   fmt.Sprintf("http://%s", l.svcCtx.MSVoteNode(nil).Address),
		WSUrl:                   wsUrl,
		ProxyFileUrl:            proxyFileUrl,
		MSVideoPlayAddressTypes: constants.VideoPlayAddressTypes,
		RtmpPort:                l.svcCtx.Config.SevBase.MediaServerRtmpPort,
		PProfFileDir:            l.svcCtx.Config.PProfFileDir,
		PProf: []tps.OptionItem{
			{Title: l.svcCtx.Config.PProf.BackendApiName, Value: l.svcCtx.Config.PProf.BackendApiPort},
			{Title: l.svcCtx.Config.PProf.DbRpcName, Value: l.svcCtx.Config.PProf.DbRpcPort},
			{Title: l.svcCtx.Config.PProf.VssName, Value: l.svcCtx.Config.PProf.VssPort},
			{Title: l.svcCtx.Config.PProf.WebName, Value: l.svcCtx.Config.PProf.WebPort},
			{Title: l.svcCtx.Config.PProf.CronName, Value: l.svcCtx.Config.PProf.CronPort},
			{Title: l.svcCtx.Config.PProf.MediaServerName, Value: l.svcCtx.Config.PProf.MediaServerPort},
		},
		AlarmTypes:          alarms.AlarmTypes,
		EventTypes:          alarms.EventTypes,
		AlarmMethods:        alarms.AlarmMethods,
		AlarmPriorities:     alarms.AlarmPriorities,
		CascadeSipProtocols: cascade.ProtocolMaps,
		SipPort:             l.svcCtx.Config.Sip.Port,
		CascadeSipPort:      l.svcCtx.Config.Sip.CascadeSipPort,
		ApiDocDir:           constants.API_DOC_DIR,
		Referer:             remoteRes.Referer,
		BuildTime:           l.svcCtx.BuildTime,
		TMapKey:             l.svcCtx.Config.TMapKey,
	}
}
