package http

import (
	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/logic/http/base"
	"skeyevss/core/app/sev/vss/internal/logic/http/gbs"
	"skeyevss/core/app/sev/vss/internal/logic/http/ms"
	"skeyevss/core/app/sev/vss/internal/logic/http/notify"
	"skeyevss/core/app/sev/vss/internal/logic/http/onvif"
	"skeyevss/core/app/sev/vss/internal/logic/http/video"
	"skeyevss/core/app/sev/vss/internal/types"
)

func RegisterApiHandlers(svcCtx *types.ServiceContext, router *gin.RouterGroup) {
	// 服务状态获取
	router.GET(base.StatusLogic.Path(), newHandler(svcCtx, base.StatusLogic))
	// 设备录像
	router.POST(base.VVideosLogic.Path(), newHandlerWithParams[types.QueryVideoRecordsReq](svcCtx, base.VVideosLogic))
	// 生成ws token
	router.POST(base.VWSTokenLogic.Path(), newHandlerWithParams[types.WSTokenReq](svcCtx, base.VWSTokenLogic))

	// video----------------------------------------
	// 获取视频播放地址
	router.POST(video.VStreamPlayLogic.Path(), newHandlerWithParams[types.VideoStreamReq](svcCtx, video.VStreamPlayLogic))
	// 停止视频播放
	router.POST(video.VStreamStopLogic.Path(), newHandlerWithParams[types.VideoStreamStopReq](svcCtx, video.VStreamStopLogic))
	// 获取视频留信息
	router.GET(video.StreamInfoLogic.Path(), newHandler(svcCtx, video.StreamInfoLogic))

	// media server----------------------------------------
	// 获取所有流组信息
	router.POST(ms.VAllGroupsLogic.Path(), newHandlerWithParams[types.DCReq](svcCtx, ms.VAllGroupsLogic))
	// 按录像流名称列表查询服务录像
	router.POST(ms.VQueryRecordByNamesLogic.Path(), newHandlerWithParams[types.MsQueryRecordByNameReq](svcCtx, ms.VQueryRecordByNamesLogic))
	// 获取流媒体服务简略配置信息
	router.POST(ms.VGetConfigLogic.Path(), newHandlerWithParams[types.MsGetConfigReq](svcCtx, ms.VGetConfigLogic))
	// 设置流媒体服务重要配置参数并重启服务
	router.POST(ms.VReloadLogic.Path(), newHandlerWithParams[types.MsReloadReq](svcCtx, ms.VReloadLogic))

	// gbs ----------------------------------------
	// 发送catalog请求
	router.GET(gbs.CatalogLogic.Path(), newHandler(svcCtx, gbs.CatalogLogic))
	// 拉流请求 直播/回放
	router.GET(gbs.InviteLogic.Path(), newHandler(svcCtx, gbs.InviteLogic))
	// 停止视频
	router.GET(gbs.StopStreamLogic.Path(), newHandler(svcCtx, gbs.StopStreamLogic))
	// 视频回放控制
	router.POST(gbs.VPlaybackControlLogic.Path(), newHandlerWithParams[types.VideoPlaybackControlReq](svcCtx, gbs.VPlaybackControlLogic))
	// 发送订阅
	router.POST(gbs.VSubscriptionLogic.Path(), newHandlerWithParams[types.SubscriptionReq](svcCtx, gbs.VSubscriptionLogic))

	// onvif ----------------------------------------
	// 探测设备
	router.GET(onvif.DiscoverLogic.Path(), newHandler(svcCtx, onvif.DiscoverLogic))
	// 获取设备信息
	router.POST(onvif.VDeviceInfoLogic.Path(), newHandlerWithParams[types.OnvifDeviceInfoReq](svcCtx, onvif.VDeviceInfoLogic))
	// 获取设备 通道信息
	router.POST(onvif.VDeviceProfilesLogic.Path(), newHandlerWithParams[types.OnvifDeviceInfoReq](svcCtx, onvif.VDeviceProfilesLogic))

	// notify ----------------------------------------
	// RTMP开始推流通知(当前做为上级,下级(设备)给当前推流)
	router.POST(notify.VOnPubStartLogic.Path(), newHandlerWithParams[types.NotifyStreamReq](svcCtx, notify.VOnPubStartLogic))
	// RTMP停止推流通知(当前做为上级,下级(设备)给当前推流)
	router.POST(notify.VOnPubStopLogic.Path(), newHandlerWithParams[types.NotifyStreamReq](svcCtx, notify.VOnPubStopLogic))
	// RTMP开始推流通知(当前作为下级(设备),给上级推流)
	router.POST(notify.VOnPushStartLogic.Path(), newHandlerWithParams[types.NotifyStreamReq](svcCtx, notify.VOnPushStartLogic))
	// RTMP停止推流通知(当前作为下级(设备),给上级推流)
	router.POST(notify.VOnPushStopLogic.Path(), newHandlerWithParams[types.NotifyStreamReq](svcCtx, notify.VOnPushStopLogic))
	// 开始拉流通知
	router.POST(notify.VOnRelayPullStartLogic.Path(), newHandlerWithParams[types.NotifyStreamReq](svcCtx, notify.VOnRelayPullStartLogic))
	// 停止拉流通知
	router.POST(notify.VOnRelayPullStopLogic.Path(), newHandlerWithParams[types.NotifyStreamReq](svcCtx, notify.VOnRelayPullStopLogic))
	// 有RTMP推流连接建立的事件通知
	router.POST(notify.VOnRtmpConnectLogic.Path(), newHandlerWithParams[types.NotifyRtmpConnectReq](svcCtx, notify.VOnRtmpConnectLogic))
	// RTMP推流通知
	router.POST(notify.VOnSubStartLogic.Path(), newHandlerWithParams[types.NotifyStreamReq](svcCtx, notify.VOnSubStartLogic))
	router.POST(notify.VOnSubStopLogic.Path(), newHandlerWithParams[types.NotifyStreamReq](svcCtx, notify.VOnSubStopLogic))

	router.POST(notify.VOnReportStatLogic.Path(), newHandlerWithParams[types.NotifyStreamReq](svcCtx, notify.VOnReportStatLogic))
	router.POST(notify.VOnHlsMakeTsLogic.Path(), newHandlerWithParams[types.NotifyStreamReq](svcCtx, notify.VOnHlsMakeTsLogic))
	router.POST(notify.VOnServerStartLogic.Path(), newHandlerWithParams[types.NotifyStreamReq](svcCtx, notify.VOnServerStartLogic))
	router.POST(notify.VOnReportFrameInfoLogic.Path(), newHandlerWithParams[types.NotifyStreamReq](svcCtx, notify.VOnReportFrameInfoLogic))
	router.POST(notify.VOnUpdateLogic.Path(), newHandlerWithParams[types.NotifyStreamReq](svcCtx, notify.VOnUpdateLogic))
}
