package gbs_sip

import (
	"context"

	gosip "github.com/ghettovoice/gosip/sip"

	"skeyevss/core/app/sev/vss/internal/logic/ws"
	"skeyevss/core/app/sev/vss/internal/pkg/common"
	sip2 "skeyevss/core/app/sev/vss/internal/pkg/sip"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/sdp"
)

var _ types.SipReceiveHandleLogic[*InviteLogic] = (*InviteLogic)(nil)

type InviteLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	req    *types.Request
	tx     gosip.ServerTransaction
}

func (l *InviteLogic) New(ctx context.Context, svcCtx *types.ServiceContext, req *types.Request, tx gosip.ServerTransaction) *InviteLogic {
	return &InviteLogic{
		svcCtx: svcCtx,
		ctx:    ctx,
		req:    req,
		tx:     tx,
	}
}

func (l *InviteLogic) DO() *types.Response {
	sdpInfo, err := sdp.ParseString(l.req.Original.Body())
	if err != nil {
		return &types.Response{Error: types.NewErr(err.Error())}
	}

	if len(sdpInfo.Media) <= 0 {
		return &types.Response{Error: types.NewErr("非法请求")}
	}

	// 检测请求合法性
	var (
		all        = l.svcCtx.TalkSipData.Keys()
		talkSipKey = ""
	)
	for _, item := range all {
		if item == l.req.ID {
			talkSipKey = item
		}
	}

	if talkSipKey == "" {
		return &types.Response{
			Code:  types.StatusUnauthorized,
			Error: types.NewErr("非法请求"),
		}
	}

	// 本地端口
	var usablePort = common.UsablePort(l.svcCtx)
	if usablePort <= 0 {
		ws.BGBSSendTalkPubError(l.svcCtx, talkSipKey, "可用端口获取失败")
		return &types.Response{
			Code:  types.StatusUnauthorized,
			Error: types.NewErr("可用端口获取失败"),
		}
	}

	// 回复invite 200OK
	if err := sip2.NewGBSSender(l.svcCtx, l.req, l.req.ID).InviteSDPResponse(l.tx, sdpInfo, usablePort); err != nil {
		return &types.Response{Error: types.NewErr(err.Error())}
	}

	// 设置rtp链接信息
	if err := common.SetTalkRtpConnInfo(l.svcCtx, sdpInfo, talkSipKey, int(usablePort)); err != nil {
		ws.BGBSSendTalkPubError(l.svcCtx, talkSipKey, err.Error())
		return &types.Response{
			Code:  types.StatusUnauthorized,
			Error: types.NewErr(err.Error()),
		}
	}

	return &types.Response{Ignore: true}
}
