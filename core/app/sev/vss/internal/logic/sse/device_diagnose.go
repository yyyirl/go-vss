// @Title        设备诊断
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package sse

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ghettovoice/gosip/sip"
	"github.com/use-go/onvif/device"
	"github.com/use-go/onvif/sdk"

	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/app/sev/vss/internal/pkg/onvif"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/app/sev/vss/internal/types/ptz"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/devices"
	"skeyevss/core/repositories/models/dictionaries"
)

type SSEDeviceDiagnosesReq struct {
	Type           string `json:"type" form:"type" path:"type" validate:"required"`
	DeviceUniqueId string `json:"deviceUniqueId" form:"deviceUniqueId" path:"deviceUniqueId" validate:"required"`
}

var (
	_ types.SSEHandleLogic[*DeviceDiagnose, *SSEDeviceDiagnosesReq] = (*DeviceDiagnose)(nil)

	DeviceDiagnosesType = "device_diagnose"

	VDeviceDiagnoses = new(DeviceDiagnose)
)

type DeviceDiagnose struct {
	ctx         context.Context
	svcCtx      *types.ServiceContext
	messageChan chan *types.SSEResponse
}

func (l *DeviceDiagnose) New(ctx context.Context, svcCtx *types.ServiceContext, messageChan chan *types.SSEResponse) *DeviceDiagnose {
	return &DeviceDiagnose{
		ctx:         ctx,
		svcCtx:      svcCtx,
		messageChan: messageChan,
	}
}

func (l *DeviceDiagnose) GetType() string {
	return DeviceDiagnosesType
}

func (l *DeviceDiagnose) end(message string) {
	l.messageChan <- &types.SSEResponse{
		Data: &types.DeviceDiagnosesResp{
			Title: "诊断结果",
			Records: []*types.DeviceDiagnosesItem{
				{Line: message},
			},
		},
	}

	l.messageChan <- &types.SSEResponse{
		Data: &types.DeviceDiagnosesResp{
			Title: "诊断结束",
			Done:  true,
		},
	}
}

func (l *DeviceDiagnose) DO(req *SSEDeviceDiagnosesReq) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 获取设备信息
	res, err := response.NewRpcToHttpResp[*deviceservice.Response, *devices.Item]().Parse(
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

	// 链路状态诊断 ------------------------------------------------------------
	var records = []*types.DeviceDiagnosesItem{
		{Title: "设备ID", Value: res.Data.DeviceUniqueId},
		{Title: "设备地址", Value: res.Data.Address},
		// {Title: "设备账号", Value: res.Data.Username},
		// {Title: "设备密码", Value: res.Data.Password},
	}

	if res.Data.Online == 1 {
		records = append(records, &types.DeviceDiagnosesItem{Title: "链路状态", Value: "在线", Color: "rgba(24, 144, 255, .9)"})
	} else {
		records = append(records, &types.DeviceDiagnosesItem{Title: "链路状态", Value: "下线", Color: "rgba(215, 26, 27, .8)"})
	}
	l.messageChan <- &types.SSEResponse{
		Data: &types.DeviceDiagnosesResp{
			Title:   "链路状态诊断",
			Records: records,
		},
	}

	if res.Data.Online != 1 {
		l.end("设备未上线")
		return
	}

	time.Sleep(300 * time.Millisecond)

	// 链路状态诊断 ------------------------------------------------------------
	records = []*types.DeviceDiagnosesItem{}
	for _, item := range l.svcCtx.DictionaryMap[dictionaries.UniqueIdDeviceManufacturer].Children {
		if uint64(item.ID) == res.Data.ManufacturerId {
			records = append(records, &types.DeviceDiagnosesItem{Title: "厂商", Value: item.Name})
		}
	}
	if res.Data.ModelVersion != "" {
		records = append(records, &types.DeviceDiagnosesItem{Title: "型号", Value: res.Data.ModelVersion})
	}

	l.messageChan <- &types.SSEResponse{
		Data: &types.DeviceDiagnosesResp{
			Title:   "设备信息",
			Records: records,
		},
	}
	time.Sleep(300 * time.Millisecond)

	// 设备能力分析 ------------------------------------------------------------
	switch res.Data.AccessProtocol {
	case devices.AccessProtocol_1, devices.AccessProtocol_2: // 流媒体源 RTMP推流
		var records []*types.DeviceDiagnosesItem
		if res.Data.OnlineAt > 0 {
			records = append(records, &types.DeviceDiagnosesItem{Title: "上线时间", Value: functions.NewTimer().FormatTimestamp(int64(res.Data.OnlineAt), "")})
		}

		if res.Data.OfflineAt > 0 {
			records = append(records, &types.DeviceDiagnosesItem{Title: "下线时间", Value: functions.NewTimer().FormatTimestamp(int64(res.Data.OfflineAt), "")})
		}

		if len(records) > 0 {
			l.messageChan <- &types.SSEResponse{
				Data: &types.DeviceDiagnosesResp{
					Title:   "设备交互",
					Records: records,
				},
			}
			time.Sleep(300 * time.Millisecond)
		}

	case devices.AccessProtocol_3: // ONVIF协议
		addrRes, err := functions.ExtractBaseURL(res.Data.Address)
		if err != nil {
			l.messageChan <- &types.SSEResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("设备地址解析错误, err: %s", err)), localization.M0010),
			}
			return
		}

		dev, err := onvif.GetDev(
			&onvif.GetDevParams{
				Username: res.Data.Username,
				Password: res.Data.Password,
				IP:       addrRes.IP,
				Port:     addrRes.Port,
			},
		)
		if err != nil {
			l.messageChan <- &types.SSEResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("client 初始化失败, err: %s", err)), localization.M0010),
			}
			return
		}

		// res, err := dev.CallMethod(ptz.GetCapabilities{})
		getCapabilitiesReq := device.GetCapabilities{Category: "All"}
		res, err := dev.CallMethod(getCapabilitiesReq)
		if err != nil {
			l.messageChan <- &types.SSEResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("设备能力获取失败, err: %s", err)), localization.M0010),
			}
			return
		}

		var reply ptz.CapabilitiesEnvelopeResp
		// 读取响应
		err = sdk.ReadAndParse(context.Background(), res, &reply, "GetPresets")
		if err != nil {
			l.messageChan <- &types.SSEResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("设备能力获取数据解析失败, err: %s", err)), localization.M0010),
			}
			return
		}

		records = []*types.DeviceDiagnosesItem{}
		if reply.Body.GetCapabilitiesResponse.Capabilities.Analytics.RuleSupport {
			records = append(records, &types.DeviceDiagnosesItem{Title: "分析能力", Value: "有", Color: "#7ADAA5"})
		} else {
			records = append(records, &types.DeviceDiagnosesItem{Title: "分析能力", Value: "无", Color: "rgba(24, 144, 255, .9)"})
		}

		if reply.Body.GetCapabilitiesResponse.Capabilities.PTZ.XAddr != "" {
			records = append(records, &types.DeviceDiagnosesItem{Title: "PTZ能力", Value: "有", Color: "#7ADAA5"})
		} else {
			records = append(records, &types.DeviceDiagnosesItem{Title: "PTZ能力", Value: "无", Color: "rgba(24, 144, 255, .9)"})
		}

		if reply.Body.GetCapabilitiesResponse.Capabilities.Media.XAddr != "" {
			records = append(records, &types.DeviceDiagnosesItem{Title: "媒体能力", Value: "有", Color: "#7ADAA5"})
		} else {
			records = append(records, &types.DeviceDiagnosesItem{Title: "媒体能力", Value: "无", Color: "rgba(24, 144, 255, .9)"})
		}

		if reply.Body.GetCapabilitiesResponse.Capabilities.Media.StreamingCapabilities.RTPMulticast {
			records = append(records, &types.DeviceDiagnosesItem{Title: "RTPMulticast", Value: "有", Color: "#7ADAA5"})
		} else {
			records = append(records, &types.DeviceDiagnosesItem{Title: "RTPMulticast", Value: "无", Color: "rgba(24, 144, 255, .9)"})
		}

		if reply.Body.GetCapabilitiesResponse.Capabilities.Media.StreamingCapabilities.RTP_TCP {
			records = append(records, &types.DeviceDiagnosesItem{Title: "RTP_TCP", Value: "有", Color: "#7ADAA5"})
		} else {
			records = append(records, &types.DeviceDiagnosesItem{Title: "RTP_TCP", Value: "无", Color: "rgba(24, 144, 255, .9)"})
		}

		if reply.Body.GetCapabilitiesResponse.Capabilities.Media.StreamingCapabilities.RTP_RTSP_TCP {
			records = append(records, &types.DeviceDiagnosesItem{Title: "RTP_RTSP_TCP", Value: "有", Color: "#7ADAA5"})
		} else {
			records = append(records, &types.DeviceDiagnosesItem{Title: "RTP_RTSP_TCP", Value: "无", Color: "rgba(24, 144, 255, .9)"})
		}

		records = append(records, &types.DeviceDiagnosesItem{
			Title: "最大通道数量",
			Value: strconv.Itoa(reply.Body.GetCapabilitiesResponse.Capabilities.Media.Extension.ProfileCapabilities.MaximumNumberOfProfiles),
			Color: "#7ADAA5",
		})

		if reply.Body.GetCapabilitiesResponse.Capabilities.Events.WSPullPointSupport {
			records = append(records, &types.DeviceDiagnosesItem{Title: "事件能力", Value: "有", Color: "#7ADAA5"})
		} else {
			records = append(records, &types.DeviceDiagnosesItem{Title: "事件能力", Value: "无", Color: "rgba(24, 144, 255, .9)"})
		}

		if reply.Body.GetCapabilitiesResponse.Capabilities.Events.WSSubscriptionPolicySupport {
			records = append(records, &types.DeviceDiagnosesItem{Title: "WSS订阅策略", Value: "有", Color: "#7ADAA5"})
		} else {
			records = append(records, &types.DeviceDiagnosesItem{Title: "WSS订阅策略", Value: "无", Color: "rgba(24, 144, 255, .9)"})
		}

		if reply.Body.GetCapabilitiesResponse.Capabilities.Events.WSPullPointSupport {
			records = append(records, &types.DeviceDiagnosesItem{Title: "WSS通知(WSPullPointSupport)", Value: "有", Color: "#7ADAA5"})
		} else {
			records = append(records, &types.DeviceDiagnosesItem{Title: "WSS通知(WSPullPointSupport)", Value: "无", Color: "rgba(24, 144, 255, .9)"})
		}

		if reply.Body.GetCapabilitiesResponse.Capabilities.Events.WSPausableSubscriptionManagerInterfaceSupport {
			records = append(records, &types.DeviceDiagnosesItem{Title: "WSPausableSubscriptionManagerInterfaceSupport", Value: "有", Color: "#7ADAA5"})
		} else {
			records = append(records, &types.DeviceDiagnosesItem{Title: "WSPausableSubscriptionManagerInterfaceSupport", Value: "无", Color: "rgba(24, 144, 255, .9)"})
		}

		l.messageChan <- &types.SSEResponse{
			Data: &types.DeviceDiagnosesResp{
				Title:   "设备能力分析",
				Records: records,
			},
		}
		time.Sleep(300 * time.Millisecond)

	case devices.AccessProtocol_4: // GB28181协议
		var records []*types.DeviceDiagnosesItem
		if res.Data.RegisterAt > 0 {
			records = append(records, &types.DeviceDiagnosesItem{Title: "最后一次注册时间", Value: functions.NewTimer().FormatTimestamp(int64(res.Data.RegisterAt), "")})
		}

		if res.Data.KeepaliveAt > 0 {
			records = append(records, &types.DeviceDiagnosesItem{Title: "最后一次心跳时间", Value: functions.NewTimer().FormatTimestamp(int64(res.Data.KeepaliveAt), "")})
		}

		if res.Data.OnlineAt > 0 {
			records = append(records, &types.DeviceDiagnosesItem{Title: "上线时间", Value: functions.NewTimer().FormatTimestamp(int64(res.Data.OnlineAt), "")})
		}

		if res.Data.OfflineAt > 0 {
			records = append(records, &types.DeviceDiagnosesItem{Title: "下线时间", Value: functions.NewTimer().FormatTimestamp(int64(res.Data.OfflineAt), "")})
		}

		if len(records) > 0 {
			l.messageChan <- &types.SSEResponse{
				Data: &types.DeviceDiagnosesResp{
					Title:   "设备交互时间",
					Records: records,
				},
			}
			time.Sleep(300 * time.Millisecond)
		}

		if l.svcCtx.Config.Sip.UsePassword {
			sipReqRes, ok := l.svcCtx.SipCatalogLoopMap.Get(res.Data.DeviceUniqueId)
			if !ok {
				l.end("设备未注册")
				return
			}

			if len(sipReqRes.Req.Authorization) >= 1 {
				authenticateHeader, ok := sipReqRes.Req.Authorization[0].(*sip.GenericHeader)
				if !ok {
					l.end("设备鉴权信息获取失败")
					return
				}

				var auth = sip.AuthFromValue(authenticateHeader.Contents)
				auth.SetPassword(l.svcCtx.Config.Sip.Password).SetMethod(string(sip.REGISTER))
				if auth.CalcResponse() != auth.Response() {
					l.end("设备账号密码与服务器不一致")
					return
				}
			} else {
				l.end("设备鉴权信息获取失败, 等待下一次注册")
			}
		}

	case devices.AccessProtocol_5: // EHOME协议
	}

	l.end("本次诊断各项指标均在正常范围内，未发现异常状况。受检设备目前状态正常，建议保持良好的网络连接。")
	l.messageChan <- &types.SSEResponse{
		Done: true,
	}
}
