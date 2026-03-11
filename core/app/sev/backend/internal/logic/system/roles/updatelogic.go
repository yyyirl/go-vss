package roles

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/common/opt"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/system-operation-logs"
)

type UpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateLogic) Update(req *orm.ReqParams) *response.HttpErr {
	// 日志记录
	opt.NewSystemOperationLogs(l.svcCtx.RpcClients).Make(l.ctx, systemOperationLogs.Types[systemOperationLogs.TypeRoleUpldate], req)

	_, err := response.NewRpcToHttpResp[*backendservice.Response, bool]().Parse(
		func() (*backendservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(req)
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Backend.RoleUpdate(l.ctx, data)
		},
	)
	if err == nil {
		l.svcCtx.AuthSet <- struct{}{}
	}

	return err
}
