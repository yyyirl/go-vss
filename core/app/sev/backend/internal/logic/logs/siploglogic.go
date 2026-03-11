package logs

import (
	"context"
	"errors"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type SipLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSipLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SipLogLogic {
	return &SipLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SipLogLogic) SipLog(req *types.FileNameQuery) (interface{}, *response.HttpErr) {
	var lines []string
	if req.FileName != "def" {
		viewer, err := functions.NewFileTailViewer(
			path.Join(
				l.svcCtx.Config.SipLogPath,
				strings.Replace(strings.Replace(req.FileName, "+", "/", -1), l.svcCtx.Config.SipLogPath, "", -1),
			),
			req.PageSize,
		)
		if err != nil {
			return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010)
		}

		lines, err = viewer.GetFileLines(req.Page, false)
		if err != nil {
			if !errors.Is(err, functions.ErrNoMoreContent) {
				return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010)
			}
		}
	}

	trees, err := functions.FileTrees(l.svcCtx.Config.SipLogPath, 0)
	if err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010)
	}

	var files = logFilesTidy(trees.Children, l.svcCtx.Config.SipLogPath)
	sort.Slice(files, func(i, j int) bool {
		ti, _ := time.Parse("2006-01-02", strings.TrimSuffix(files[i].Name, ".log"))
		tj, _ := time.Parse("2006-01-02", strings.TrimSuffix(files[j].Name, ".log"))

		return ti.After(tj)
	})

	return map[string]interface{}{
		"lines": lines,
		"files": files,
		"dir":   l.svcCtx.Config.SipLogPath,
	}, nil
}
