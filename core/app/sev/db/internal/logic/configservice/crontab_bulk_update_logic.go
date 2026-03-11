package configservicelogic

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/crontab"
)

type CrontabBulkUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCrontabBulkUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CrontabBulkUpdateLogic {
	return &CrontabBulkUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 任务批量更新
func (l *CrontabBulkUpdateLogic) CrontabBulkUpdate(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	if len(params.BulkUpdates) <= 0 {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	if err := response.NewMakeRpcRetErr(l.svcCtx.CrontabModel.BulkUpdate(
		crontab.ColumnUniqueId,
		[]string{crontab.ColumnUpdatedAt, crontab.ColumnLogs},
		params.BulkUpdates,
	), 2); err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return &db.Response{
		Data:    []byte(strconv.FormatBool(true)),
		License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string),
	}, nil
}
