// @Title        文件下载
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package sse

import (
	"context"
	"time"

	"skeyevss/core/app/sev/vss/internal/types"
)

type (
	SSESevStateReq struct {
		Type string `json:"type" form:"type" path:"type" validate:"required"`
	}

	SSESevStateItem struct {
		Title string `json:"title"`
		Num   int    `json:"num"`
	}
)

var (
	_ types.SSEHandleLogic[*SevState, *SSESevStateReq] = (*SevState)(nil)

	SevStateType = "sev_state"

	VSevState = new(SevState)
)

type SevState struct {
	ctx         context.Context
	svcCtx      *types.ServiceContext
	messageChan chan *types.SSEResponse
}

func (l *SevState) New(ctx context.Context, svcCtx *types.ServiceContext, messageChan chan *types.SSEResponse) *SevState {
	return &SevState{
		ctx:         ctx,
		svcCtx:      svcCtx,
		messageChan: messageChan,
	}
}

func (l *SevState) GetType() string {
	return SevStateType
}

func (l *SevState) DO(_ *SSESevStateReq) {
	var ticker = time.NewTicker(time.Second * 1)
	for {
		select {
		case <-l.ctx.Done():
			return

		case <-ticker.C:
			l.messageChan <- &types.SSEResponse{
				Data: []SSESevStateItem{
					{
						Title: "文件下载任务数量",
						Num:   l.svcCtx.DownloadManager.TaskNum(),
					},
					{
						Title: "文件下载任务数量",
						Num:   l.svcCtx.DownloadManager.ClientNum(),
					},
					{
						Title: "catalog请求节流器",
						Num:   l.svcCtx.SipCatalogLoopMap.Len(),
					},
					{
						Title: "心跳检测节流器",
						Num:   l.svcCtx.SipHeartbeatLoopMap.Len(),
					},
					{
						Title: "invite请求限制防止并发击穿信令",
						Num:   l.svcCtx.InviteRequestState.Size(),
					},
					{
						Title: "流是否存在",
						Num:   l.svcCtx.PubStreamExistsState.Size(),
					},
					{
						Title: "GBS SN",
						Num:   l.svcCtx.SipGBSSNMap.Len(),
					},
					{
						Title: "GBC SN",
						Num:   l.svcCtx.SipGBCSNMap.Len(),
					},
					{
						Title: "GBS Ack",
						Num:   l.svcCtx.AckRequestMap.Len(),
					},
					{
						Title: "设置设备在线状态节流器",
						Num:   l.svcCtx.DeviceOnlineStateUpdateMap.Len(),
					},
					{
						Title: "Onvif设备探测",
						Num:   len(l.svcCtx.OnvifDiscoverDevices),
					},
					{
						Title: "MS记录",
						Num:   len(l.svcCtx.MediaServerRecords),
					},
					{
						Title: "设备级联",
						Num:   len(l.svcCtx.CascadeRecords),
					},
					{
						Title: "级联注册",
						Num:   l.svcCtx.CascadeRegister.Len(),
					},
					{
						Title: "级联心跳",
						Num:   l.svcCtx.CascadeKeepaliveCounter.Len(),
					},
					{
						Title: "级联注册Executing",
						Num:   l.svcCtx.CascadeRegisterExecuting.Len(),
					},
					{
						Title: "GBC Invite请求限流",
						Num:   l.svcCtx.GBCInviteReqMaps.Len(),
					},
					{
						Title: "GBC获取设备录像标志",
						Num:   l.svcCtx.GBCRecordInfoSendMaps.Len(),
					},
					{
						Title: "Websocket链接数量",
						Num:   l.svcCtx.WSClientCache.Len(),
					},
				},
			}
		}
	}
}
