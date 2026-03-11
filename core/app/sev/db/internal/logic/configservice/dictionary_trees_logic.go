package configservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/categories"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/dictionaries"
)

type DictionaryTreesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDictionaryTreesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictionaryTreesLogic {
	return &DictionaryTreesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DictionaryTreesLogic) DictionaryTrees(_ *db.EmptyRequest) (*db.Response, error) {
	list, queryErr := l.svcCtx.DictionariesModel.List(&orm.ReqParams{
		All: true,
	})
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

	return response.NewRpcResp[*db.Response]().Make(
		categories.New[int, *dictionaries.Item]().Conv(
			records,
			func(item *dictionaries.Item) *categories.Item[int, *dictionaries.Item] {
				return &categories.Item[int, *dictionaries.Item]{
					ID: int(item.ID), Pid: int(item.ParentId), Name: item.Name, Raw: item,
				}
			},
		).Trees,
		3,
		func(data []byte) *db.Response {
			return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
		},
	)
}
