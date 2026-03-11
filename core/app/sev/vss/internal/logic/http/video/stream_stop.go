package video

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/app/sev/vss/internal/logic/http/gbs"
	"skeyevss/core/app/sev/vss/internal/pkg/ms"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/common/stream"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/devices"
)

var (
	_ types.HttpRHandleLogic[*StreamStopLogic, types.VideoStreamStopReq] = (*StreamStopLogic)(nil)

	VStreamStopLogic = new(StreamStopLogic)
)

type StreamStopLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *StreamStopLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *StreamStopLogic {
	return &StreamStopLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *StreamStopLogic) Path() string {
	return "/video/stream/stop"
}

func (l *StreamStopLogic) DO(req types.VideoStreamStopReq) *types.HttpResponse {
	var msAddress = ms.New(l.ctx, l.svcCtx).VoteNode([]uint64{req.ID}).Address
	var streamNames []string
	if req.StreamName != "" {
		streamNames = append(streamNames, req.StreamName)
	} else {
		streamNames = req.StreamNames
	}

	if len(streamNames) <= 0 {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("streamName 不能为空"), localization.M0001),
		}
	}

	for _, streamName := range streamNames {
		// l.svcCtx.PlaybackControlMap.Remove(streamName)

		res, err1 := stream.New().Parse(streamName)
		if err1 != nil {
			return &types.HttpResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Err(err1), localization.MR1003),
			}
		}

		// 获取设备信息
		deviceRes, err := response.NewRpcToHttpResp[*backendservice.Response, *devices.Item]().Parse(
			func() (*backendservice.Response, error) {
				data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
					Conditions: []*orm.ConditionItem{
						{Column: devices.ColumnDeviceUniqueId, Value: res.Device},
					},
				})
				if err != nil {
					return nil, err
				}

				return l.svcCtx.RpcClients.Device.DeviceRow(l.ctx, data)
			},
		)
		if err != nil {
			return &types.HttpResponse{Err: err}
		}

		var Type = "pull"
		if deviceRes.Data.AccessProtocol == devices.AccessProtocol_2 { // RTMP推流
			groupInDetailResp, _, err := ms.New(l.ctx, l.svcCtx).GetStreamGroup(fmt.Sprintf("http://%s/api", msAddress), streamName)
			if err != nil {
				return &types.HttpResponse{
					Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.MR1003),
				}
			}

			if groupInDetailResp == nil || groupInDetailResp.Pub == nil {
				return &types.HttpResponse{
					Err: response.MakeError(response.NewHttpRespMessage().Str("会话不存在"), localization.MR1003),
				}
			}

			if _, _, err := ms.New(l.ctx, l.svcCtx).KickSession(fmt.Sprintf("http://%s/api", msAddress), streamName, groupInDetailResp.Pub.SessionID); err != nil {
				return &types.HttpResponse{
					Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.MR1003),
				}
			}

			return nil
		} else if deviceRes.Data.AccessProtocol == devices.AccessProtocol_4 { // GB28181推流
			Type = "pub" // 国标

			// 发送停止BYE请求 停止国标推流
			if resp := gbs.StopStreamLogic.New(l.ctx, l.c, l.svcCtx).StopStream(streamName, "0"); resp != nil && resp.Err != nil {
				return &types.HttpResponse{
					Err: response.MakeError(response.NewHttpRespMessage().Str(resp.Err.Error), localization.MR1003),
				}
			}
			l.svcCtx.AckRequestMap.Remove(streamName)
		} else { // rtsp拉流
			// 停止流媒体Pub/Pull
			if err := ms.New(l.ctx, l.svcCtx).StopMSStream(msAddress, streamName, Type); err != nil {
				return &types.HttpResponse{
					Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.MR1003),
				}
			}
		}
	}

	return nil
}
