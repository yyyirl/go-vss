package internalCall

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type DeviceControlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeviceControlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeviceControlLogic {
	return &DeviceControlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeviceControlLogic) DeviceControl(req map[string]interface{}, Type string) *response.HttpErr {
	var rq = l.svcCtx.RemoteReq(l.ctx)
	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode, Referer: rq.Referer}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/device-control?t="+Type, rq.VssHttpUrlInternal),
		req,
		nil,
	); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Str(fmt.Sprintf("设备控制, err: %s", err)), localization.M0010)
	}

	return nil
}
