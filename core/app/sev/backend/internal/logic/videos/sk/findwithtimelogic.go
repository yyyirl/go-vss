package sk

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

type FindWithTimeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFindWithTimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindWithTimeLogic {
	return &FindWithTimeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FindWithTimeLogic) FindWithTime(req *types.QueryVideoRecordsReq) (string, *response.HttpErr) {
	videoRes, _, msAddress, err := records(l.ctx, l.svcCtx, req)
	if err != nil {
		return "", err
	}

	if len(videoRes.Data) <= 0 {
		return "", nil
	}

	var (
		start = req.StartDate / 1000
		end   = req.EndDate / 1000
	)
	for _, item := range videoRes.Data {
		var tmp = item.Date / 1000
		if !(int64(start) <= tmp && int64(end) >= tmp+item.Duration/1000) {
			continue
		}

		return msAddress + "/playback/" + strings.TrimPrefix(item.RecordPath, "/"), nil
	}

	return "", response.MakeError(response.NewHttpRespMessage().Str("未获取到相关记录"), localization.M0010)
}
