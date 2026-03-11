package video

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/pkg/ms"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

var (
	_ types.HttpEHandleLogic[*streamInfoLogic] = (*streamInfoLogic)(nil)

	StreamInfoLogic = new(streamInfoLogic)
)

type streamInfoLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *streamInfoLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *streamInfoLogic {
	return &streamInfoLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *streamInfoLogic) Path() string {
	return "/video/stream/:msId/:streamName"
}

func (l *streamInfoLogic) DO() *types.HttpResponse {
	var streamName = l.c.Param("streamName")
	if streamName == "" {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("stream name 不能为空"), localization.M0001),
		}
	}

	var msId = l.c.Param("msId")
	if msId == "" {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("ms id 不能为空"), localization.M0001),
		}
	}

	id, err := strconv.Atoi(msId)
	if err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("ms id 类型错误"), localization.M0001),
		}
	}

	groupInDetailResp, _, err := ms.New(l.ctx, l.svcCtx).GetStreamGroup(fmt.Sprintf("http://%s/api", ms.New(l.ctx, l.svcCtx).VoteNode([]uint64{uint64(id)}).Address), streamName)
	if err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00300),
		}
	}

	if groupInDetailResp == nil || groupInDetailResp.Pub == nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("会话不存在"), localization.MR1003),
		}
	}

	return &types.HttpResponse{Data: groupInDetailResp}
}
