package dictionaries

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/app/sev/db/client/configservice"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type CheckUniqueIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckUniqueIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckUniqueIdLogic {
	return &CheckUniqueIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckUniqueIdLogic) CheckUniqueId(req *types.UniqueIdQuery) (interface{}, *response.HttpErr) {
	res, err := response.NewRpcToHttpResp[*configservice.Response, bool]().Parse(
		func() (*configservice.Response, error) {
			return l.svcCtx.RpcClients.Config.DictionaryCheckUniqueId(
				l.ctx,
				&configservice.UniqueIdReq{UniqueId: req.UniqueId},
			)
		},
	)
	if err != nil {
		return false, err
	}

	return functions.ByteToBool(res.Res.Data), nil
}
