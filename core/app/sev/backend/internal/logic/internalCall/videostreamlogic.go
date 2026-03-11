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

type VideoStreamLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVideoStreamLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VideoStreamLogic {
	return &VideoStreamLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VideoStreamLogic) VideoStream(req map[string]interface{}) (interface{}, *response.HttpErr) {
	var (
		rq  = l.svcCtx.RemoteReq(l.ctx)
		res response.HttpResp[map[string]interface{}]
	)
	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode, Referer: rq.Referer}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/video/stream", rq.VssHttpUrlInternal),
		req,
		&res,
	); err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(fmt.Sprintf("视频播放获取播放地址, [%s] [%s] err: %s", rq.VssHttpUrlInternal, rq.Referer, err)), localization.M0010)
	}

	if res.Error != "" {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(res.Error), localization.M0010)
	}

	return res.Data, nil
}
