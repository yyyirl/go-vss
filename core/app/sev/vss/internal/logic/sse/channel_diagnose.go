// @Title        通道诊断
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package sse

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/app/sev/vss/internal/logic/gbs_proc"
	"skeyevss/core/app/sev/vss/internal/logic/http/gbs"
	"skeyevss/core/app/sev/vss/internal/pkg/ms"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/common/stream"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
	"skeyevss/core/repositories/models/devices"
)

type SSEChannelDiagnosesReq struct {
	Type            string `json:"type" form:"type" path:"type" validate:"required"`
	ChannelUniqueId string `json:"channelUniqueId" form:"channelUniqueId" path:"channelUniqueId" validate:"required"`
	DeviceUniqueId  string `json:"deviceUniqueId" form:"deviceUniqueId" path:"deviceUniqueId" validate:"required"`
}

var (
	_ types.SSEHandleLogic[*ChannelDiagnose, *SSEChannelDiagnosesReq] = (*ChannelDiagnose)(nil)

	ChannelDiagnosesType = "channel_diagnose"

	VChannelDiagnoses = new(ChannelDiagnose)
)

type ChannelDiagnose struct {
	ctx         context.Context
	svcCtx      *types.ServiceContext
	messageChan chan *types.SSEResponse
}

func (l *ChannelDiagnose) New(ctx context.Context, svcCtx *types.ServiceContext, messageChan chan *types.SSEResponse) *ChannelDiagnose {
	return &ChannelDiagnose{
		ctx:         ctx,
		svcCtx:      svcCtx,
		messageChan: messageChan,
	}
}

func (l *ChannelDiagnose) GetType() string {
	return ChannelDiagnosesType
}

func (l *ChannelDiagnose) done() {
	l.messageChan <- &types.SSEResponse{
		Data: &types.DeviceDiagnosesResp{
			Title: "诊断结束",
			Done:  true,
		},
	}

	l.messageChan <- &types.SSEResponse{
		Done: true,
	}
}

func (l *ChannelDiagnose) DO(req *SSEChannelDiagnosesReq) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 获取设备信息
	deviceRes, err := response.NewRpcToHttpResp[*deviceservice.Response, *devices.Item]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: devices.ColumnDeviceUniqueId, Value: req.DeviceUniqueId},
				},
			})
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.DeviceRow(ctx, data)
		},
	)
	if err != nil {
		l.messageChan <- &types.SSEResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("设备获取失败, err: %s", err)), localization.M0010),
		}
		return
	}

	// 获取通道信息
	channelRes, err := response.NewRpcToHttpResp[*deviceservice.Response, *channels.Item]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: channels.ColumnUniqueId, Value: req.ChannelUniqueId},
					{Column: channels.ColumnDeviceUniqueId, Value: req.DeviceUniqueId},
				},
			})
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.ChannelRowFind(ctx, data)
		},
	)
	if err != nil {
		l.messageChan <- &types.SSEResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("设备获取失败, err: %s", err)), localization.M0010),
		}
		return
	}

	defer l.done()

	switch deviceRes.Data.AccessProtocol {
	case devices.AccessProtocol_1, devices.AccessProtocol_3: // 流媒体源 ONVIF协议
		var (
			records = []*types.DeviceDiagnosesItem{
				{Title: "通道ID", Value: channelRes.Data.UniqueId},
			}
			streamUrl = channelRes.Data.StreamUrl
		)
		if deviceRes.Data.AccessProtocol == devices.AccessProtocol_1 {
			streamUrl = deviceRes.Data.StreamUrl
		}

		res, err := ms.NewRTSPChecker(5 * time.Second).Check(streamUrl)
		if err != nil {
			records = append(records, &types.DeviceDiagnosesItem{
				Line: "检测失败: " + err.Error(),
			})
		}

		if res != nil {
			if res.IsOnline {
				records = append(records, &types.DeviceDiagnosesItem{
					Title: "流状态",
					Value: "在线",
				})
			} else {
				if res.Error != nil {
					records = append(records, &types.DeviceDiagnosesItem{
						Title: "错误信息",
						Line:  res.Error.Error(),
						Color: "rgba(215, 26, 27, .8)",
					})
				} else {
					records = append(records, &types.DeviceDiagnosesItem{
						Title: "流状态",
						Value: "离线",
					})
				}
			}

			if res.MediaInfo != "" {
				records = append(records, &types.DeviceDiagnosesItem{
					Line: res.MediaInfo,
				})
			}
		}

		records = append(records, &types.DeviceDiagnosesItem{
			Line: "流地址: " + streamUrl,
		})

		l.messageChan <- &types.SSEResponse{
			Data: &types.DeviceDiagnosesResp{
				Title:   "视频诊断",
				Records: records,
			},
		}

	case devices.AccessProtocol_2: // RTMP推流
		l.rtmp(deviceRes.Data, channelRes.Data)

	case devices.AccessProtocol_4: // GB28181协议
		l.gb28181(deviceRes.Data, channelRes.Data)

	case devices.AccessProtocol_5: // EHOME协议

	}
}

func (l *ChannelDiagnose) rtmp(deviceItem *devices.Item, channelItem *channels.Item) {
	var (
		streamName = stream.New().Produce(deviceItem.DeviceUniqueId, channelItem.UniqueId, stream.PlayTypePlay)
		streamUrl  = fmt.Sprintf("%s/%s", strings.Trim(deviceItem.StreamUrl, "/"), streamName)
		records    = []*types.DeviceDiagnosesItem{
			{Title: "通道ID", Value: channelItem.UniqueId},
			{Line: "流地址: " + streamUrl},
		}
	)

	if channelItem.Online == 1 {
		records = append(
			records,
			&types.DeviceDiagnosesItem{Title: "在线状态", Value: "在线", Color: "rgba(24, 144, 255, .9)"},
			&types.DeviceDiagnosesItem{Title: "上线时间", Value: functions.NewTimer().FormatTimestamp(int64(channelItem.OnlineAt), "")},
		)
	} else {
		records = append(records, &types.DeviceDiagnosesItem{Title: "在线状态", Value: "下线", Color: "rgba(215, 26, 27, .8)"})
	}

	l.messageChan <- &types.SSEResponse{
		Data: &types.DeviceDiagnosesResp{
			Title:   "视频诊断",
			Records: records,
		},
	}

	// 拉流 检测
	streamRes, err := ms.New(l.ctx, l.svcCtx).StreamInfo(deviceItem)
	if err != nil {
		l.messageChan <- &types.SSEResponse{
			Data: &types.DeviceDiagnosesResp{
				Title: "诊断结果",
				Records: []*types.DeviceDiagnosesItem{
					{Line: "media server 信息获取失败"},
				},
			},
		}
		return
	}

	res, _, _ := ms.New(l.ctx, l.svcCtx).GetStreamGroup(streamRes.MediaServerUrl, streamName)
	if res == nil {
		l.messageChan <- &types.SSEResponse{
			Data: &types.DeviceDiagnosesResp{
				Title: "诊断结果",
				Records: []*types.DeviceDiagnosesItem{
					{Line: "拉流信息获取失败"},
				},
			},
		}
		return
	}

	if res.Pub == nil {
		l.messageChan <- &types.SSEResponse{
			Data: &types.DeviceDiagnosesResp{
				Title: "诊断结果",
				Records: []*types.DeviceDiagnosesItem{
					{Line: "未开始拉流"},
				},
			},
		}
		return
	}

	ctx1, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var (
		ticker               = time.NewTicker(1 * time.Second)
		readStreamState      = false
		readBytesSum    uint = 0
	)
	defer ticker.Stop()

Loop1:
	for {
		select {
		case <-ctx1.Done():
			break Loop1

		case <-ticker.C:
			if readStreamState {
				continue
			}

			readStreamState = true
			if res, _, _ := ms.New(l.ctx, l.svcCtx).GetStreamGroup(streamRes.MediaServerUrl, streamName); res != nil {
				if res.Pub != nil {
					readBytesSum += res.Pub.ReadBytesSum
				}
			}
			readStreamState = false
		}
	}

	l.messageChan <- &types.SSEResponse{
		Data: &types.DeviceDiagnosesResp{
			Title: "视频诊断",
			Records: []*types.DeviceDiagnosesItem{
				{Title: "Session ID", Value: res.Pub.SessionID},
				{Title: "Protocol", Value: res.Pub.Protocol},
				{Title: "Type", Value: res.Pub.BaseType},
				{Title: "推流地址", Value: res.Pub.RemoteAddr},
				{Title: "推流时间", Value: res.Pub.StartTime},
				{Title: "读取字节数", Value: functions.ByteSize(uint64(readBytesSum))},
			},
		},
	}
}

func (l *ChannelDiagnose) gb28181(deviceItem *devices.Item, channelItem *channels.Item) {
	var records = []*types.DeviceDiagnosesItem{
		{Title: "通道id", Value: channelItem.UniqueId},
	}
	if channelItem.Online == 1 {
		records = append(
			records,
			&types.DeviceDiagnosesItem{Title: "在线状态", Value: "在线", Color: "rgba(24, 144, 255, .9)"},
			&types.DeviceDiagnosesItem{Title: "上线时间", Value: functions.NewTimer().FormatTimestamp(int64(channelItem.OnlineAt), "")},
		)
	} else {
		records = append(records, &types.DeviceDiagnosesItem{Title: "在线状态", Value: "下线", Color: "rgba(215, 26, 27, .8)"})
	}

	records = append(records, &types.DeviceDiagnosesItem{Title: "注册时间", Value: functions.NewTimer().FormatTimestamp(int64(deviceItem.RegisterAt), "")})
	if v, ok := l.svcCtx.SipHeartbeatLoopMap.Get(channelItem.DeviceUniqueId); ok {
		records = append(records, &types.DeviceDiagnosesItem{Title: "上一次心跳时间", Value: functions.NewTimer().FormatTimestamp(v.Now, "")})
	}

	l.messageChan <- &types.SSEResponse{
		Data: &types.DeviceDiagnosesResp{
			Title:   "通道信息",
			Records: records,
		},
	}

	if channelItem.Online != 1 {
		l.messageChan <- &types.SSEResponse{
			Data: &types.DeviceDiagnosesResp{
				Title: "诊断结果",
				Records: []*types.DeviceDiagnosesItem{
					{Line: "通道不在线"},
				},
			},
		}
		return
	}

	// 拉流 检测
	streamRes, err := ms.New(l.ctx, l.svcCtx).StreamInfo(deviceItem)
	if err != nil {
		l.messageChan <- &types.SSEResponse{
			Data: &types.DeviceDiagnosesResp{
				Title: "诊断结果",
				Records: []*types.DeviceDiagnosesItem{
					{Line: "拉流信息获取失败"},
				},
			},
		}
		return
	}

	sipReqRes, ok := l.svcCtx.SipCatalogLoopMap.Get(channelItem.DeviceUniqueId)
	if !ok {
		l.messageChan <- &types.SSEResponse{
			Data: &types.DeviceDiagnosesResp{
				Title: "诊断结果",
				Records: []*types.DeviceDiagnosesItem{
					{Line: "设备未上线", Color: "rgba(215, 26, 27, .8)"},
				},
			},
		}
		return
	}

	var (
		stepInfo   = &types.StepRecord{Message: make(chan *types.StepRecordMessage, 10)}
		streamName = stream.New().Produce(channelItem.DeviceUniqueId, channelItem.UniqueId, stream.PlayTypePlay)
	)
	defer func() {
		close(stepInfo.Message)
	}()

	var msIP = streamRes.MSNode.InternalIP
	if l.svcCtx.Config.Sip.UseExternalWan {
		msIP = streamRes.MSNode.ExtIP
	}

	if err := gbs_proc.NewSendLogic(l.svcCtx, func(name string) {
		if err := recover(); err != nil {
			functions.LogError(fmt.Sprintf("Sip Interval [%s] Recover [%s] \nStack: %s", name, err, string(debug.Stack())))
		}
	}).VideoLiveInvite(
		&types.SipVideoLiveInviteMessage{
			StreamPort:        streamRes.StreamPort,
			MediaTransMode:    streamRes.TransportProtocol.MediaTransMode,
			MediaServerUrl:    streamRes.MediaServerUrl,
			MediaServerIP:     msIP,
			MediaServerPort:   streamRes.MSNode.HttpPort,
			StreamName:        streamName,
			PlayType:          stream.PlayTypePlay,
			DeviceUniqueId:    channelItem.DeviceUniqueId,
			MediaProtocolMode: streamRes.TransportProtocol.MediaProtocolMode,
			ChannelUniqueId:   channelItem.UniqueId,
			Req:               sipReqRes.Req,
			TransportProtocol: streamRes.TransportProtocol,
			Data:              deviceItem,
			StepInfo:          stepInfo,
		},
	); err != nil {
		l.messageChan <- &types.SSEResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010),
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	l.messageChan <- &types.SSEResponse{
		Data: &types.DeviceDiagnosesResp{
			Title: "检测流状态",
		},
	}

	var sendSipRecords []*types.DeviceDiagnosesItem
Loop:
	for {
		select {
		case <-ctx.Done():
			time.Sleep(300 * time.Millisecond)
			l.messageChan <- &types.SSEResponse{Data: &types.DeviceDiagnosesResp{Line: "诊断超时"}}
			break Loop

		case v := <-stepInfo.Message:
			if v.Done {
				// 发送bye请求，并停止流媒体RTP PUB
				if resp := gbs.StopStreamLogic.New(l.ctx, nil, l.svcCtx).StopStream(streamName, "0"); resp != nil && resp.Err != nil {
					functions.LogError("bye 请求发送失败")
				}
				time.Sleep(300 * time.Millisecond)
				l.messageChan <- &types.SSEResponse{Data: &types.DeviceDiagnosesResp{Line: ""}}
				break Loop
			}

			if v.SipContent != nil {
				sendSipRecords = append(
					sendSipRecords,
					&types.DeviceDiagnosesItem{
						Title: v.SipContent.Type,
						Line:  v.SipContent.Content,
					},
				)
				continue
			}

			if v.Error != nil {
				time.Sleep(300 * time.Millisecond)
				l.messageChan <- &types.SSEResponse{Data: &types.DeviceDiagnosesResp{Line: "诊断错误[003]"}}
				time.Sleep(300 * time.Millisecond)
				l.messageChan <- &types.SSEResponse{
					Err: response.MakeError(response.NewHttpRespMessage().Err(v.Error), localization.M0010),
				}

				break Loop
			}

			time.Sleep(300 * time.Millisecond)
			l.messageChan <- &types.SSEResponse{Data: &types.DeviceDiagnosesResp{Line: v.Message}}
		}
	}

	if len(sendSipRecords) > 0 {
		l.messageChan <- &types.SSEResponse{
			Data: &types.DeviceDiagnosesResp{
				Title: "信令交互",
			},
		}

		for _, item := range sendSipRecords {
			l.messageChan <- &types.SSEResponse{
				Data: &types.DeviceDiagnosesResp{
					Line: item.Line,
				},
			}
		}
	}

	var mediaProtocolMode = "UDP"
	if streamRes.TransportProtocol.MediaProtocolMode == 1 {
		mediaProtocolMode = "TCP"
	}

	var records1 = []*types.DeviceDiagnosesItem{
		{
			Title: "流媒体服务名称",
			Value: streamRes.MSNode.Name,
		},
		{
			Title: "流媒体服务IP",
			Value: streamRes.MSNode.IP,
		},
		{
			Title: "收流协议",
			Value: mediaProtocolMode,
		},
	}

	ctx1, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var (
		ticker               = time.NewTicker(1 * time.Second)
		readStreamState      = false
		readBytesSum    uint = 0
	)
	defer ticker.Stop()

Loop1:
	for {
		select {
		case <-ctx1.Done():
			break Loop1

		case <-ticker.C:
			if readStreamState {
				continue
			}

			readStreamState = true
			if res, _, _ := ms.New(l.ctx, l.svcCtx).GetStreamGroup(streamRes.MediaServerUrl, streamName); res != nil {
				if res.Pub != nil {
					readBytesSum += res.Pub.ReadBytesSum
				}
			}
			readStreamState = false
		}
	}

	if readBytesSum > 0 {
		records1 = append(records1, &types.DeviceDiagnosesItem{
			Title: "读取字节数",
			Value: functions.ByteSize(uint64(readBytesSum)),
		})
	}

	l.messageChan <- &types.SSEResponse{
		Data: &types.DeviceDiagnosesResp{
			Title:   "诊断视频流",
			Records: records1,
		},
	}
}
