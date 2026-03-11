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

type MSReloadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMSReloadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MSReloadLogic {
	return &MSReloadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MSReloadLogic) MSReload(req *types.SetSMSConfigReq) *response.HttpErr {
	var rq = l.svcCtx.RemoteReq(l.ctx)
	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode, Referer: rq.Referer}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/ms/reload", rq.VssHttpUrlInternal),
		map[string]interface{}{
			"ip":     req.IP,
			"port":   req.Port,
			"reboot": req.Reboot,
			"delay":  req.Delay,
			"config": req.Config,
		},
		nil,
	); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Str(fmt.Sprintf("设置流媒体服务重要配置参数并重启服务, err: %s", err)), localization.M0010)
	}

	return nil
}
