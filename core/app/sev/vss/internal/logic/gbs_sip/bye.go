package gbs_sip

import (
	"context"

	gosip "github.com/ghettovoice/gosip/sip"

	"skeyevss/core/app/sev/vss/internal/types"
)

var _ types.SipReceiveHandleLogic[*ByeLogic] = (*ByeLogic)(nil)

type ByeLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	req    *types.Request
	tx     gosip.ServerTransaction
}

func (l *ByeLogic) New(ctx context.Context, svcCtx *types.ServiceContext, req *types.Request, tx gosip.ServerTransaction) *ByeLogic {
	return &ByeLogic{
		svcCtx: svcCtx,
		ctx:    ctx,
		req:    req,
		tx:     tx,
	}
}

func (l *ByeLogic) DO() *types.Response {
	// if v, ok := l.req.Original.CallID(); ok {
	// 	var (
	// 		callID = v.String()
	// 		maps   = l.svcCtx.TalkSipData.All()
	// 	)
	// 	for key, item := range maps {
	// 		if callID != item.CallID {
	// 			continue
	// 		}
	//
	// 		ws.RGBSTalkAudioStop(l.svcCtx, key)
	// 		break
	// 	}
	// }

	return nil
}
