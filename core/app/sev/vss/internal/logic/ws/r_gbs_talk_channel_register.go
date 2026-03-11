package ws

import (
	"context"

	"skeyevss/core/app/sev/vss/internal/types"
)

const GbsTalkChannelRegisterKey = "gbs-talk-channel-register"

type RGBSTalkSipChannelRegisterLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	client *types.WSClient
}

func NewRGBSTalkChannelRegister(ctx context.Context, svcCtx *types.ServiceContext, client *types.WSClient) *RGBSTalkSipChannelRegisterLogic {
	return &RGBSTalkSipChannelRegisterLogic{ctx: ctx, svcCtx: svcCtx, client: client}
}

func (l *RGBSTalkSipChannelRegisterLogic) Do(req *types.WSGBSTalkChannelRegister) *types.WSResponse {
	var message = "sip注册成功"
	if req.Offline {
		l.client.SipTalkActivateKey = ""
		message = "sip注销成功"
	} else {
		l.client.SipTalkActivateKey = req.DeviceUniqueId
	}
	return &types.WSResponse{
		Message: message,
	}
}
