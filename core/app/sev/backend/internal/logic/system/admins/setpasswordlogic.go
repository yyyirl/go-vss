package admins

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/common/opt"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/system-operation-logs"
)

type SetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetPasswordLogic {
	return &SetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetPasswordLogic) SetPassword(req *types.SetPasswordReq) *response.HttpErr {
	// 日志记录
	opt.NewSystemOperationLogs(l.svcCtx.RpcClients).Make(l.ctx, systemOperationLogs.Types[systemOperationLogs.TypeAdminUpdate], req)

	_, err := response.NewRpcToHttpResp[*backendservice.Response, string]().Parse(
		func() (*backendservice.Response, error) {

			return l.svcCtx.RpcClients.Backend.AdminPassword(l.ctx, &backendservice.AdminPasswordReq{
				Password:    req.Password,
				OldPassword: req.OldPassword,
			})
		},
	)
	if err == nil {
		l.svcCtx.AuthSet <- struct{}{}
	}

	return err
}
