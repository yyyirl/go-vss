package dictionaries

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/configservice"
	"skeyevss/core/pkg/categories"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/dictionaries"
)

type TreesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTreesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TreesLogic {
	return &TreesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TreesLogic) Trees() (interface{}, *response.HttpErr) {
	res, err := response.NewRpcToHttpResp[*configservice.Response, []*categories.Item[int, *dictionaries.Item]]().Parse(
		func() (*configservice.Response, error) {
			return l.svcCtx.RpcClients.Config.DictionaryTrees(l.ctx, &configservice.EmptyRequest{})
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}
