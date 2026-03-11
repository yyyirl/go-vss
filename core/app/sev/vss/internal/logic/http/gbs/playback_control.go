package gbs

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/pkg/sip"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/common/stream"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

var (
	_ types.HttpRHandleLogic[*PlaybackControlLogic, types.VideoPlaybackControlReq] = (*PlaybackControlLogic)(nil)

	VPlaybackControlLogic = new(PlaybackControlLogic)
)

type PlaybackControlLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *PlaybackControlLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *PlaybackControlLogic {
	return &PlaybackControlLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *PlaybackControlLogic) Path() string {
	return "/gbs/playback-control"
}

func (l *PlaybackControlLogic) DO(req types.VideoPlaybackControlReq) *types.HttpResponse {
	if req.StreamName == "" {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("参数错误"), localization.M0001),
		}
	}

	streamRes, err := stream.New().Parse(req.StreamName)
	if err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("参数错误"), localization.M0001),
		}
	}

	res, ok := l.svcCtx.AckRequestMap.Get(req.StreamName)
	if !ok {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("请求参数获取失败"), localization.M0010),
		}
	}

	if _, err = sip.NewGBSSender(l.svcCtx, res.Req, res.ChannelUniqueId).PlaybackControl(res.SendData, streamRes, req); err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.MR1007),
		}
	}

	return nil
}
