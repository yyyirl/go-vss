package internalCall

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type MSQueryRecordByNamesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMSQueryRecordByNamesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MSQueryRecordByNamesLogic {
	return &MSQueryRecordByNamesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MSQueryRecordByNamesLogic) MSQueryRecordByNames(req *types.MsQueryRecordByNameReq) (interface{}, *response.HttpErr) {
	var (
		res response.HttpResp[[]map[string]interface{}]
		rq  = l.svcCtx.RemoteReq(l.ctx)
	)
	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode, Referer: rq.Referer}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/ms/query_record_by_names", rq.VssHttpUrlInternal),
		map[string]interface{}{
			"streamNames":     req.StreamNames,
			"recordType":      req.RecordType,
			"channelUniqueId": req.ChannelUniqueId,
			"deviceUniqueId":  req.DeviceUniqueId,
		},
		&res,
	); err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(fmt.Sprintf("按录像流名称列表查询服务录像, err: %s", err)), localization.M0010)
	}

	if res.Error != "" {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(res.Error), localization.M0010)
	}

	return res.Data, nil
}
