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

type PresetPointsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPresetPointsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PresetPointsLogic {
	return &PresetPointsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PresetPointsLogic) PresetPoints(req map[string]interface{}) (interface{}, *response.HttpErr) {
	var (
		res response.HttpResp[map[string]interface{}]
		rq  = l.svcCtx.RemoteReq(l.ctx)
	)
	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode, Referer: rq.Referer}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/preset-points", rq.VssHttpUrlInternal),
		req,
		&res,
	); err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(fmt.Sprintf("获取预设位, err: %s", err)), localization.M0010)
	}

	if res.Error != "" {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(res.Error), localization.M0010)
	}

	return res.Data, nil
}
