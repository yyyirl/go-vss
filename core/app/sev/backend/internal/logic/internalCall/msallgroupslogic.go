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

type MSAllGroupsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMSAllGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MSAllGroupsLogic {
	return &MSAllGroupsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MSAllGroupsLogic) MSAllGroups(req *types.DCReq) (interface{}, *response.HttpErr) {
	var (
		res response.HttpResp[[]map[string]interface{}]
		rq  = l.svcCtx.RemoteReq(l.ctx)
	)
	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode, Referer: rq.Referer}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/ms/all_groups", rq.VssHttpUrlInternal),
		map[string]interface{}{
			"channelUniqueId": req.ChannelUniqueId,
			"deviceUniqueId":  req.DeviceUniqueId,
			"msID":            req.MsID,
		},
		&res,
	); err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(fmt.Sprintf("获取所有流组信息获取失败, err: %s", err)), localization.M0010)
	}

	if res.Error != "" {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(res.Error), localization.M0010)
	}

	return res.Data, nil
}
