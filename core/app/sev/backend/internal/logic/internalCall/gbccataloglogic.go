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

type GBCCatalogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGBCCatalogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GBCCatalogLogic {
	return &GBCCatalogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GBCCatalogLogic) GBCCatalog(req map[string]interface{}) *response.HttpErr {
	var rq = l.svcCtx.RemoteReq(l.ctx)
	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode, Referer: rq.Referer}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/gbc/catalog", rq.VssHttpUrlInternal),
		req,
		nil,
	); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Str(fmt.Sprintf("gbc发送catalog请求, err: %s", err)), localization.M0010)
	}

	return nil
}
