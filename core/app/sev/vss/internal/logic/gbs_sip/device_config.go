package gbs_sip

import (
	"context"

	"skeyevss/core/app/sev/vss/internal/types"

	gosip "github.com/ghettovoice/gosip/sip"
)

var _ types.SipReceiveHandleLogic[*DeviceConfigLogic] = (*DeviceConfigLogic)(nil)

type DeviceConfigLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	req    *types.Request
	tx     gosip.ServerTransaction
}

func (l *DeviceConfigLogic) New(ctx context.Context, svcCtx *types.ServiceContext, req *types.Request, tx gosip.ServerTransaction) *DeviceConfigLogic {
	return &DeviceConfigLogic{
		svcCtx: svcCtx,
		ctx:    ctx,
		req:    req,
		tx:     tx,
	}
}

func (l *DeviceConfigLogic) DO() *types.Response {
	return &types.Response{}
}
