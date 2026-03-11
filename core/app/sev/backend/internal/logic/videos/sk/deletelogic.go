package sk

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/common/opt"
	"skeyevss/core/common/stream"
	ctypes "skeyevss/core/common/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	systemOperationLogs "skeyevss/core/repositories/models/system-operation-logs"
)

type DeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLogic {
	return &DeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteLogic) Delete(req *orm.ReqParams) *response.HttpErr {
	// 日志记录
	opt.NewSystemOperationLogs(l.svcCtx.RpcClients).Make(l.ctx, systemOperationLogs.Types[systemOperationLogs.TypeVideoSKDelete], req)

	if len(req.Conditions) <= 0 {
		return response.MakeError(response.NewHttpRespMessage().Str("conditions 不能为空"), localization.M0001)
	}

	conditions, err := pickConditions(req)
	if err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001)
	}

	// 获取设备,通道信息
	res, err1 := response.NewRpcToHttpResp[*deviceservice.Response, *ctypes.DeviceChannel]().Parse(
		func() (*deviceservice.Response, error) {
			return l.svcCtx.RpcClients.Device.DeviceChannel(l.ctx, &deviceservice.DeviceChannelReq{
				ChannelUniqueId: conditions.ChannelUniqueId,
				DeviceUniqueId:  conditions.DeviceUniqueId,
			})
		},
	)
	if err1 != nil {
		return err1
	}

	if res == nil || res.Data == nil {
		return response.MakeError(response.NewHttpRespMessage().Str("数据获取失败, 返回值为空[2]"), localization.MR1008)
	}

	var (
		streamName = stream.New().Produce(res.Data.Device.DeviceUniqueId, res.Data.Channel.UniqueId, stream.PlayTypePlay)
		Type       = "files"
	)
	if req.All {
		Type = "all"
	}

	if conditions.Start > 0 && conditions.End > 0 {
		Type = "times"
	}

	var files []string
	for _, item := range conditions.Filenames {
		_, filename := filepath.Split(item)
		files = append(files, filename)
	}

	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode}).HttpPostJsonResJson(
		fmt.Sprintf("http://%s/api/record/delete", l.svcCtx.MSVoteNode(res.Data.Device.MSIds).Address),
		map[string]interface{}{
			"stream_name": streamName,
			"files":       files,
			"type":        Type,
			"start_time":  conditions.Start,
			"end_time":    conditions.End,
			"record_type": 0,
		},
		nil,
	); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00271)
	}

	return nil
}
