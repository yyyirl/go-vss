package configservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/dictionaries"
)

type DictionariesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDictionariesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictionariesLogic {
	return &DictionariesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DictionariesLogic) Dictionaries(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	// 获取总数
	count, queryErr := l.svcCtx.DictionariesModel.Count(params)
	if queryErr != nil {
		return nil, response.NewMakeRpcRetErr(queryErr, 2)
	}

	if count <= 0 {
		return response.NewRpcResp[*db.Response]().Make(response.NewListResp[[]*dictionaries.Item]().Empty(), 3, func(data []byte) *db.Response {
			return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
		})
	}

	// 获取列表
	list, queryErr := l.svcCtx.DictionariesModel.List(params)
	if queryErr != nil {
		return nil, response.NewMakeRpcRetErr(queryErr, 2)
	}

	var records []*dictionaries.Item
	for _, item := range list {
		v, err := item.ConvToItem()
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		records = append(records, v)
	}

	return response.NewRpcResp[*db.Response]().Make(&response.ListResp[[]*dictionaries.Item]{
		List:  records,
		Count: count,
	}, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
