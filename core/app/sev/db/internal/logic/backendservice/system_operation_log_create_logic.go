package backendservicelogic

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/system-operation-logs"
)

type SystemOperationLogCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSystemOperationLogCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SystemOperationLogCreateLogic {
	return &SystemOperationLogCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SystemOperationLogCreateLogic) SystemOperationLogCreate(in *db.MapReq) (*db.Response, error) {
	// 操作日志创建
	record, err := systemOperationLogs.NewItem().MapToModel(in.Data.AsMap())
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	if _, ok := systemOperationLogs.Types[record.Type]; !ok {
		record.Type = systemOperationLogs.Types[systemOperationLogs.Known]
	}

	res, err := l.svcCtx.SystemOperationLogsModel.Add(record.SystemOperationLogs.Correction(orm.ActionInsert).(systemOperationLogs.SystemOperationLogs))
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return &db.Response{
		Data:    []byte(strconv.Itoa(int(res.ID))),
		License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string),
	}, nil
}
