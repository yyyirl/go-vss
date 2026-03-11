package gbs_sip

import (
	"context"

	gosip "github.com/ghettovoice/gosip/sip"

	"skeyevss/core/app/sev/vss/internal/pkg/sip"
	"skeyevss/core/app/sev/vss/internal/types"
)

var _ types.SipReceiveHandleLogic[*MediaStatusLogic] = (*MediaStatusLogic)(nil)

type MediaStatusLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	req    *types.Request
	tx     gosip.ServerTransaction
}

func (l *MediaStatusLogic) New(ctx context.Context, svcCtx *types.ServiceContext, req *types.Request, tx gosip.ServerTransaction) *MediaStatusLogic {
	return &MediaStatusLogic{
		svcCtx: svcCtx,
		ctx:    ctx,
		req:    req,
		tx:     tx,
	}
}

func (l *MediaStatusLogic) DO() *types.Response {
	data, err := sip.NewParser[types.SipMessageMediaStatus]().ToData(l.req.Original)
	if err != nil {
		return &types.Response{Error: types.NewErr(err.Error())}
	}

	if data.NotifyType == 121 {
		// 发送bye 请求
		if _, ok := l.svcCtx.AckRequestMap.Get(data.DeviceID); ok {
			l.svcCtx.SipSendBye <- &types.SipByeMessage{
				StreamName: data.DeviceID,
			}
		}
	}

	return nil
}
