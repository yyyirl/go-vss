package logs

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/system-operation-logs"
)

type SystemOperationLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSystemOperationLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SystemOperationLogsLogic {
	return &SystemOperationLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SystemOperationLogsLogic) SystemOperationLogs(req *orm.ReqParams) (interface{}, *response.HttpErr) {
	res, err := response.NewRpcToHttpResp[*backendservice.Response, *response.ListResp[[]*systemOperationLogs.Item]]().Parse(
		func() (*backendservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(req)
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Backend.SystemOperationLogs(l.ctx, data)
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}
