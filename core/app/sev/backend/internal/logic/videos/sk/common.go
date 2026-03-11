// @Title        common
// @Description  main
// @Create       yiyiyi 2025/8/19 09:29

package sk

import (
	"context"
	"errors"
	"fmt"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/common/stream"
	ctypes "skeyevss/core/common/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
)

type conditions struct {
	Filenames []string
	DeviceUniqueId,
	ChannelUniqueId string
	Start,
	End uint64
}

func pickConditions(req *orm.ReqParams) (*conditions, error) {
	var (
		filenames []string
		channelUniqueId,
		deviceUniqueId string
		start,
		end uint64
		ok bool
	)
	for _, item := range req.Conditions {
		if item.Column == "record_path" {
			if len(item.Values) <= 0 && item.Value == nil {
				return nil, errors.New("condition value values 不能同时为空")
			}

			if item.Value != nil {
				filename, ok := item.Value.(string)
				if !ok {
					return nil, errors.New("condition item 类型错误")
				}

				filenames = append(filenames, filename)
				continue
			}

			for _, v := range item.Values {
				filename, ok := v.(string)
				if !ok {
					return nil, errors.New("condition item 类型错误[1]")
				}
				filenames = append(filenames, filename)
			}
			continue
		}

		if item.Column == "deviceUniqueId" {
			if item.Value == nil {
				return nil, errors.New("deviceUniqueId 不能为空")
			}

			deviceUniqueId, ok = item.Value.(string)
			if !ok {
				return nil, errors.New("deviceUniqueId 类型错误")
			}
		}

		if item.Column == "channelUniqueId" {
			if item.Value == nil {
				return nil, errors.New("channelUniqueId 不能为空")
			}

			channelUniqueId, ok = item.Value.(string)
			if !ok {
				return nil, errors.New("channelUniqueId 类型错误")
			}
		}

		if item.Column == "start" {
			if item.Value == nil {
				return nil, errors.New("start 不能为空")
			}

			var err error
			start, err = functions.InterfaceToNumber[uint64](item.Value)
			if err != nil {
				return nil, err
			}
		}

		if item.Column == "end" {
			if item.Value == nil {
				return nil, errors.New("end 不能为空")
			}

			var err error
			end, err = functions.InterfaceToNumber[uint64](item.Value)
			if err != nil {
				return nil, err
			}
		}
	}

	return &conditions{
		Filenames:       filenames,
		DeviceUniqueId:  deviceUniqueId,
		ChannelUniqueId: channelUniqueId,
		Start:           start,
		End:             end,
	}, nil
}

func records(ctx context.Context, svcCtx *svc.ServiceContext, req *types.QueryVideoRecordsReq) (*ctypes.MediaResponse[[]*ListItem], *ctypes.DeviceChannel, string, *response.HttpErr) {
	if req.StartDate <= 0 || req.EndDate <= 0 || req.DeviceUniqueId == "" || req.ChannelUniqueId == "" {
		return nil, nil, "", response.MakeError(response.NewHttpRespMessage().Str("StartDate EndDate DeviceUniqueId ChannelUniqueId 不能为空"), localization.M0001)
	}

	// 获取设备,通道信息
	res, err := response.NewRpcToHttpResp[*deviceservice.Response, *ctypes.DeviceChannel]().Parse(
		func() (*deviceservice.Response, error) {
			return svcCtx.RpcClients.Device.DeviceChannel(ctx, &deviceservice.DeviceChannelReq{
				ChannelUniqueId: req.ChannelUniqueId,
				DeviceUniqueId:  req.DeviceUniqueId,
			})
		},
	)
	if err != nil {
		return nil, nil, "", err
	}

	if res == nil || res.Data == nil {
		return nil, nil, "", response.MakeError(response.NewHttpRespMessage().Str("数据获取失败, 返回值为空"), localization.MR1008)
	}

	var (
		streamName = stream.New().Produce(res.Data.Device.DeviceUniqueId, res.Data.Channel.UniqueId, stream.PlayTypePlay)
		videoRes   ctypes.MediaResponse[[]*ListItem]
		msAddress  = fmt.Sprintf("http://%s", svcCtx.MSVoteNode(res.Data.Device.MSIds).Address)
	)
	if _, err := functions.NewResty(ctx, &functions.RestyConfig{Mode: svcCtx.Config.Mode}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/record/query_daily", msAddress),
		map[string]interface{}{
			"stream_name": streamName,
			"start_time":  req.StartDate,
			"end_time":    req.EndDate,
			"record_type": 0,
		},
		&videoRes,
	); err != nil {
		return nil, nil, "", response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010)
	}

	if videoRes.Code != 10000 {
		return nil, nil, "", response.MakeError(response.NewHttpRespMessage().Str("code: %d, msg: %s", videoRes.Code, videoRes.Msg), localization.M0010)
	}

	return &videoRes, res.Data, msAddress, nil
}
