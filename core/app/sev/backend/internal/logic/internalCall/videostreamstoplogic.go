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

type VideoStreamStopLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVideoStreamStopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VideoStreamStopLogic {
	return &VideoStreamStopLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VideoStreamStopLogic) VideoStreamStop(req map[string]interface{}) *response.HttpErr {
	var rq = l.svcCtx.RemoteReq(l.ctx)
	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode, Referer: rq.Referer}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/video/stream/stop", rq.VssHttpUrlInternal),
		req,
		nil,
	); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Str(fmt.Sprintf("停止流, err: %s", err)), localization.M0010)
	}

	return nil
}
