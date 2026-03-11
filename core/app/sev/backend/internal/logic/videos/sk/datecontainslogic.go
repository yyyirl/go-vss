package sk

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/common/stream"
	ctypes "skeyevss/core/common/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type DateContainsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type DateContainsRes struct {
	Date    string `json:"date"`
	Records string `json:"records"`
}

func NewDateContainsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DateContainsLogic {
	return &DateContainsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DateContainsLogic) DateContains(req *types.QueryDateContainsReq) (interface{}, *response.HttpErr) {
	if req.Date <= 0 || req.DeviceUniqueId == "" || req.ChannelUniqueId == "" {
		return nil, response.MakeError(response.NewHttpRespMessage().Str("Date DeviceUniqueId ChannelUniqueId 不能为空"), localization.M0001)
	}

	// 获取设备,通道信息
	res, err := response.NewRpcToHttpResp[*deviceservice.Response, *ctypes.DeviceChannel]().Parse(
		func() (*deviceservice.Response, error) {
			return l.svcCtx.RpcClients.Device.DeviceChannel(l.ctx, &deviceservice.DeviceChannelReq{
				ChannelUniqueId: req.ChannelUniqueId,
				DeviceUniqueId:  req.DeviceUniqueId,
			})
		},
	)
	if err != nil {
		return 0, err
	}

	if res == nil || res.Data == nil {
		return 0, response.MakeError(response.NewHttpRespMessage().Str("数据获取失败, 返回值为空[1]"), localization.MR1008)
	}

	var (
		date            = functions.NewTimer().FormatTimestamp(int64(req.Date), functions.TimeFormatYm)
		streamName      = stream.New().Produce(res.Data.Device.DeviceUniqueId, res.Data.Channel.UniqueId, stream.PlayTypePlay)
		dateContainsRes response.HttpResp[*DateContainsRes]
	)
	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode}).HttpPostJsonResJson(
		fmt.Sprintf("http://%s/api/record/query_monthly", l.svcCtx.MSVoteNode(res.Data.Device.MSIds).Address),
		map[string]interface{}{
			"stream_name": streamName,
			"date":        date,
			"record_type": 0,
		},
		&dateContainsRes,
	); err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010)
	}

	return dateContainsRes.Data, nil
}
