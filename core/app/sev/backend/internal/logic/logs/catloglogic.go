package logs

import (
	"context"
	"errors"
	"path"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type CatLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCatLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CatLogLogic {
	return &CatLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CatLogLogic) CatLog(req *types.FileNameQuery) (interface{}, *response.HttpErr) {
	viewer, err := functions.NewFileTailViewer(
		path.Join(
			l.svcCtx.Config.LogPath,
			strings.Replace(strings.Replace(req.FileName, "+", "/", -1), l.svcCtx.Config.LogPath, "", -1),
		),
		req.PageSize,
	)
	if err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010)
	}

	lines, err := viewer.GetFileLines(req.Page, true)
	if err != nil {
		if !errors.Is(err, functions.ErrNoMoreContent) {
			return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010)
		}
	}

	return lines, nil
}
