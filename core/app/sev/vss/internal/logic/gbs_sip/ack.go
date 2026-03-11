package gbs_sip

import (
	"context"

	gosip "github.com/ghettovoice/gosip/sip"

	"skeyevss/core/app/sev/vss/internal/logic/ws"
	"skeyevss/core/app/sev/vss/internal/pkg/common"
	"skeyevss/core/app/sev/vss/internal/types"
)

var _ types.SipReceiveHandleLogic[*ACKLogic] = (*ACKLogic)(nil)

type ACKLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	req    *types.Request
	tx     gosip.ServerTransaction
}

func (l *ACKLogic) New(ctx context.Context, svcCtx *types.ServiceContext, req *types.Request, tx gosip.ServerTransaction) *ACKLogic {
	return &ACKLogic{
		svcCtx: svcCtx,
		ctx:    ctx,
		req:    req,
		tx:     tx,
	}
}

func (l *ACKLogic) DO() *types.Response {
	// invite和ack中没有发送通道id 统一默认接受到ack请求后认为 当前设备下所有通道sip状态都成功
	var maps = l.svcCtx.TalkSipData.All()
	for key, item := range maps {
		if key == l.req.ID {
			if err := common.SetTalkRtpConn(l.svcCtx, l.req.Original, key, item); err != nil {
				ws.BGBSSendTalkPubError(l.svcCtx, key, err.Error())
			}

			return nil
		}
	}

	return &types.Response{
		Code:  types.StatusForbidden,
		Error: types.NewErr("非法请求"),
	}
}
