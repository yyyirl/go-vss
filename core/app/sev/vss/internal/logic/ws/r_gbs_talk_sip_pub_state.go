package ws

import (
	"context"

	"skeyevss/core/app/sev/vss/internal/types"
)

const GbsTalkSipPubStateKey = "gbs-talk-sip-pub-state"

type RGBSTalkSipPubStateLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	client *types.WSClient
}

func NewRGBSTalkSipPubState(ctx context.Context, svcCtx *types.ServiceContext, client *types.WSClient) *RGBSTalkSipPubStateLogic {
	return &RGBSTalkSipPubStateLogic{ctx: ctx, svcCtx: svcCtx, client: client}
}

func (l *RGBSTalkSipPubStateLogic) Do(req *types.WSGBSTalkSipPub) *types.WSResponse {
	v, ok := l.svcCtx.TalkSipData.Get(req.DeviceUniqueId)
	if !ok {
		return &types.WSResponse{
			Data: makeRGBSTalkSipPubState(0),
		}
	}

	if v.Status {
		return &types.WSResponse{
			Data: makeRGBSTalkSipPubState(3),
		}
	}

	return &types.WSResponse{
		Data: makeRGBSTalkSipPubState(2),
	}
}

func makeRGBSTalkSipPubState(state int) map[string]interface{} {
	return map[string]interface{}{
		"state": state,
	}
}
