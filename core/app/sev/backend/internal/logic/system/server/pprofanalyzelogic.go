package server

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/pkg/pprof"
	"skeyevss/core/pkg/response"
)

type PProfAnalyzeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPProfAnalyzeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PProfAnalyzeLogic {
	return &PProfAnalyzeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PProfAnalyzeLogic) PProfAnalyze(req *pprof.PProfParams) (interface{}, *response.HttpErr) {
	req.Dir = l.svcCtx.Config.PProfFileDir
	req.Ctx = l.ctx
	results, filePaths := pprof.NewAnalyzePProf(req)
	return map[string]interface{}{
		"results":   results,
		"filePaths": filePaths,
	}, nil
}
