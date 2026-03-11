package admins

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
)

type ExistsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExistsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExistsLogic {
	return &ExistsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExistsLogic) Exists(req *orm.ReqParams) (bool, *response.HttpErr) {
	res, err := response.NewRpcToHttpResp[*backendservice.Response, bool]().Parse(
		func() (*backendservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(req)
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Backend.AdminCheckExists(l.ctx, data)
		},
	)
	if err != nil {
		return false, err
	}

	return res.Data, nil
}
