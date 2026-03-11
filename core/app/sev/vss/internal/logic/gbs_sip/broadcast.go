package gbs_sip

import (
	"context"

	gosip "github.com/ghettovoice/gosip/sip"

	"skeyevss/core/app/sev/vss/internal/pkg/sip"
	"skeyevss/core/app/sev/vss/internal/types"
)

var _ types.SipReceiveHandleLogic[*BroadcastLogic] = (*BroadcastLogic)(nil)

type BroadcastLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	req    *types.Request
	tx     gosip.ServerTransaction
}

func (l *BroadcastLogic) New(ctx context.Context, svcCtx *types.ServiceContext, req *types.Request, tx gosip.ServerTransaction) *BroadcastLogic {
	return &BroadcastLogic{
		svcCtx: svcCtx,
		ctx:    ctx,
		req:    req,
		tx:     tx,
	}
}

func (l *BroadcastLogic) DO() *types.Response {
	data, err := sip.NewParser[types.SipMessageBroadcast]().ToData(l.req.Original)
	if err != nil {
		return &types.Response{Error: types.NewErr(err.Error())}
	}

	var callID string
	if callId, ok := l.req.Original.CallID(); ok {
		callID = callId.String()
	}
	if data.Info != nil && data.Info.Reason != "" && callID != "" {
		// 语音对讲
		// TODO 完整版请联系作者
	}

	if data.Result == "OK" {
		return nil
	}

	return &types.Response{Error: types.NewErr("响应错误")}
}
