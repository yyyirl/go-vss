// @Title        recording
// @Description  main
// @Create       yiyiyi 2025/9/19 08:49

package videoProject

import (
	"context"
	"fmt"
	"time"

	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/common/client"
	"skeyevss/core/common/stream"
	cTypes "skeyevss/core/common/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
)

type Recording struct {
}

type Params struct {
	VssHttpTarget,
	Mode string
	Timeout       uint
	IsDownloading bool
	GetMSAddress  func(msIds []uint64) string
	StreamName    string
	PlayType      stream.PlayType
	RpcClients    *client.GRPCClients

	EndAt,
	StartAt uint64
}

func NewRecoding() *Recording {
	return &Recording{}
}

func (l *Recording) StopRecording(params *Params, maps map[uint64]*cTypes.ChannelMSRelItem) {
	for _, val := range maps {
		go func() {
			var ctx = context.Background()
			if params.Timeout <= 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), time.Duration(params.Timeout)*time.Second)
				defer cancel()
			}

			var (
				msAddress  = fmt.Sprintf("http://%s/api/record/stop", params.GetMSAddress(val.MSIds))
				streamName = stream.New().Produce(val.DeviceUniqueId, val.ChannelUniqueId, params.PlayType)
			)
			if params.StreamName != "" {
				streamName = params.StreamName
			}

			if _, err := functions.NewResty(ctx, &functions.RestyConfig{Mode: params.Mode}).HttpPostJsonResJson(
				msAddress,
				map[string]interface{}{
					"stream_name": streamName,
					"record_type": 0,
				},
				nil,
			); err != nil {
				functions.LogError("停止录像失败, err:", err.Error())
				return
			}

			if !params.IsDownloading {
				// 更新通道状态
				l.setRecordingState(params, maps, 0)
			}
		}()
	}
}

func (l *Recording) StartRecording(params *Params, maps map[uint64]*cTypes.ChannelMSRelItem) {
	for _, val := range maps {
		go func() {
			var ctx = context.Background()
			if params.Timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), time.Duration(params.Timeout)*time.Second)
				defer cancel()
			}

			var (
				msAddress     = fmt.Sprintf("http://%s/api/record/start", params.GetMSAddress(val.MSIds))
				streamName    = stream.New().Produce(val.DeviceUniqueId, val.ChannelUniqueId, params.PlayType)
				isDownloading = params.IsDownloading && params.StartAt != 0 && params.EndAt != 0
			)
			if params.StreamName != "" {
				streamName = params.StreamName
			}

			// 创建group
			{
				var (
					req = map[string]interface{}{
						"deviceUniqueId":  val.DeviceUniqueId,
						"channelUniqueId": val.ChannelUniqueId,
					}
					url = fmt.Sprintf("http://%s/api/video/stream", params.VssHttpTarget)
				)
				if isDownloading {
					req["download"] = true
					req["startAt"] = params.StartAt
					req["endAt"] = params.EndAt
					url = fmt.Sprintf("%s?streamName=%s", url, streamName)
				}

				if _, err := functions.NewResty(ctx, &functions.RestyConfig{Mode: params.Mode}).HttpPostJsonResJson(url, req, nil); err != nil {
					functions.LogError("创建播放group失败, err:", err.Error())
					return
				}
			}

			// 开始录像
			{
				var (
					recordInterval = 60
					recordType     = 0
				)
				if isDownloading {
					recordInterval = 0
					recordType = 2
				}

				if _, err := functions.NewResty(ctx, &functions.RestyConfig{Mode: params.Mode}).HttpPostJsonResJson(
					msAddress,
					map[string]interface{}{
						"stream_name":     streamName,
						"record_interval": recordInterval, // 录像分割文件时间片间隔，单位：秒，默认0录像不分割
						"record_type":     recordType,     // 录像类型：0-计划录像 1-报警录像 2-其他手动录像，默认0
						"record_format":   0,              // 录像存储格式：0-mp4 1-ts 2-flv，默认0存储MP4
					},
					nil,
				); err != nil {
					functions.LogError("开始录像失败, err:", err.Error())
					return
				}
			}

			if !params.IsDownloading {
				// 更新通道状态
				l.setRecordingState(params, maps, 1)
			}
		}()
	}
}

func (l *Recording) setRecordingState(params *Params, maps map[uint64]*cTypes.ChannelMSRelItem, val uint) {
	// 更新通道状态
	if _, err := response.NewRpcToHttpResp[*deviceservice.Response, uint64]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(params.Mode).ToPBParams(&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{
						Column: channels.ColumnID,
						Values: functions.SliceToSliceAny(functions.MapKeys(maps)),
					},
				},
				Data: []*orm.UpdateItem{
					{Column: channels.ColumnRecordingState, Value: val},
				},
			})
			if err != nil {
				return nil, err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			return params.RpcClients.Device.ChannelUpdate(ctx, data)
		},
	); err != nil {
		functions.LogError("设备通道播放状态更新失败, err: ", err.Error)
	}
}
