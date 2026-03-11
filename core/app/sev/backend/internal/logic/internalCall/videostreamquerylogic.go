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

type VideoStreamQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVideoStreamQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VideoStreamQueryLogic {
	return &VideoStreamQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VideoStreamQueryLogic) VideoStreamQuery(req *types.VideoStreamInfoReq) (interface{}, *response.HttpErr) {
	var (
		res response.HttpResp[map[string]interface{}]
		rq  = l.svcCtx.RemoteReq(l.ctx)
	)
	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode, Referer: rq.Referer}).HttpGetResJson(
		fmt.Sprintf("%s/api/video/stream/%d/%s", rq.VssHttpUrlInternal, req.ID, req.StreamName),
		nil,
		&res,
	); err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(fmt.Sprintf("获取流信息, err: %s", err)), localization.M0010)
	}

	if res.Error != "" {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(res.Error), localization.M0010)
	}

	return res.Data, nil
}
