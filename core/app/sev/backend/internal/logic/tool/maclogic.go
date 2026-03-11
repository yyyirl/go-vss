package tool

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

type MACLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMACLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MACLogic {
	return &MACLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MACLogic) MAC() (interface{}, *response.HttpErr) {
	var ip = contextx.GetCtxIP(l.ctx)
	if ip == "" {
		return nil, response.MakeError(response.NewHttpRespMessage().Str("ip获取失败"), localization.M0010)
	}

	mac, err := functions.GetMacAddr()
	if err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("mac地址获取失败 %s", err)), localization.M0010)
	}

	return mac, nil
}
