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

type GBSCatalogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGBSCatalogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GBSCatalogLogic {
	return &GBSCatalogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GBSCatalogLogic) GBSCatalog(req *types.DeviceUniqueIdReq) *response.HttpErr {
	var rq = l.svcCtx.RemoteReq(l.ctx)
	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode, Referer: rq.Referer}).HttpGet(
		fmt.Sprintf("%s/api/gbs/catalog/"+req.DeviceUniqueId, rq.VssHttpUrlInternal),
		nil,
	); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Str(fmt.Sprintf("gbs发送catalog请求, err: %s", err)), localization.M0010)
	}

	return nil
}
