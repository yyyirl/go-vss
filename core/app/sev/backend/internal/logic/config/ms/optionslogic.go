package ms

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/configservice"
	"skeyevss/core/pkg/response"
	"skeyevss/core/tps"
)

type OptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OptionsLogic {
	return &OptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OptionsLogic) Options() (interface{}, *response.HttpErr) {
	res, err := response.NewRpcToHttpResp[*configservice.Response, []*tps.OptionItem]().Parse(
		func() (*configservice.Response, error) {
			return l.svcCtx.RpcClients.Config.MsOptions(l.ctx, &configservice.EmptyRequest{})
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}
