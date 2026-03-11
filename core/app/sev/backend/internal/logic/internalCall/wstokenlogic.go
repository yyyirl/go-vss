package internalCall

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type WSTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWSTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WSTokenLogic {
	return &WSTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WSTokenLogic) WSToken() (interface{}, *response.HttpErr) {
	var (
		res response.HttpResp[string]
		rq  = l.svcCtx.RemoteReq(l.ctx)
	)
	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode, Referer: rq.Referer}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/ws-token", rq.VssHttpUrlInternal),
		map[string]interface{}{
			"id": contextx.GetCtxUserid(l.ctx),
		},
		&res,
	); err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(fmt.Sprintf("获取wstoken, err: %s", err)), localization.M0010)
	}

	if res.Error != "" {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(res.Error), localization.M0010)
	}

	return res.Data, nil
}
