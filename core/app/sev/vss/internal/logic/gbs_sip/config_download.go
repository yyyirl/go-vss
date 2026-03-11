package gbs_sip

import (
	"context"

	gosip "github.com/ghettovoice/gosip/sip"

	"skeyevss/core/app/sev/vss/internal/types"
)

var _ types.SipReceiveHandleLogic[*ConfigDownloadLogic] = (*ConfigDownloadLogic)(nil)

type ConfigDownloadLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	req    *types.Request
	tx     gosip.ServerTransaction
}

func (l *ConfigDownloadLogic) New(ctx context.Context, svcCtx *types.ServiceContext, req *types.Request, tx gosip.ServerTransaction) *ConfigDownloadLogic {
	return &ConfigDownloadLogic{
		svcCtx: svcCtx,
		ctx:    ctx,
		req:    req,
		tx:     tx,
	}
}

func (l *ConfigDownloadLogic) DO() *types.Response {
	return &types.Response{}
}
