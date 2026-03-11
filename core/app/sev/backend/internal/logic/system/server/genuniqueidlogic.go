package server

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	configClient "skeyevss/core/app/sev/db/client/configservice"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type GenUniqueIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenUniqueIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenUniqueIdLogic {
	return &GenUniqueIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenUniqueIdLogic) GenUniqueId(req *types.GenUniqueIdReq) ([]string, *response.HttpErr) {
	var (
		count     = req.Count
		uniqueIds []string
	)
	if count <= 0 {
		count = 1
	}
	switch req.Type {
	case "short":
		for i := 0; i < count; i++ {
			uniqueIds = append(uniqueIds, functions.GenerateUniqueID(8))
		}

		return uniqueIds, nil

	case "uniqueId":
		for i := 0; i < count; i++ {
			uniqueIds = append(uniqueIds, functions.UniqueId())
		}

		return uniqueIds, nil

	case "cascadeDepCode", "cascadeChannel":
		res, err := response.NewRpcToHttpResp[*configClient.Response, string]().Parse(
			func() (*configClient.Response, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				return l.svcCtx.RpcClients.Config.MaxId(ctx, &configClient.TypeReq{Type: req.Type})
			},
		)

		if err != nil {
			return nil, err
		}

		tmp, err1 := functions.NewBigNumber(res.Data)
		if err1 != nil {
			return nil, response.MakeError(response.NewHttpRespMessage().Err(err1), localization.M0002)
		}

		var id = l.svcCtx.Config.GenUniqueId.Dir
		if req.Type == "cascadeChannel" {
			id = l.svcCtx.Config.GenUniqueId.Camera
		}

		def, err1 := functions.NewBigNumber(id)
		if err1 != nil {
			return nil, response.MakeError(response.NewHttpRespMessage().Err(err1), localization.M0002)
		}

		if tmp.String() == "" || tmp.String() == "0" || tmp.LessThan(def) {
			tmp = def
		}

		for i := 0; i < count; i++ {
			tmp = tmp.AddOne()
			uniqueIds = append(uniqueIds, tmp.String())
		}
	}

	return uniqueIds, nil
}
