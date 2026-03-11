package logs

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type RunLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRunLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunLogsLogic {
	return &RunLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RunLogsLogic) RunLogs() (interface{}, *response.HttpErr) {
	trees, err := functions.FileTrees(l.svcCtx.Config.LogPath, 0)
	if err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010)
	}

	return &response.ListResp[[]*item]{
		List: logFilesTidy(trees.Children, l.svcCtx.Config.LogPath),
	}, nil
}
