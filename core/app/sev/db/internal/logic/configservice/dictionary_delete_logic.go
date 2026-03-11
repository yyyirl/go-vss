package configservicelogic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/dictionaries"
)

type DictionaryDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDictionaryDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictionaryDeleteLogic {
	return &DictionaryDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DictionaryDeleteLogic) DictionaryDelete(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	var ids []uint
	for _, item := range params.Conditions {
		if item.Column == dictionaries.ColumnId {
			if len(item.Values) > 0 {
				for _, v := range item.Values {
					id, err := functions.InterfaceToNumber[uint](v)
					if err != nil {
						return nil, response.NewMakeRpcRetErr(errors.New("id type error, "+err.Error()), 2)
					}
					ids = append(ids, id)
				}
				continue
			}

			if item.Value != nil {
				id, err := functions.InterfaceToNumber[uint](item.Value)
				if err != nil {
					return nil, response.NewMakeRpcRetErr(errors.New("id type error, "+err.Error()), 2)
				}
				ids = append(ids, id)
				continue
			}

			return nil, response.NewMakeRpcRetErr(errors.New("conditions id can be not empty "), 2)
		}
	}

	if len(ids) <= 0 {
		return nil, response.NewMakeRpcRetErr(errors.New("conditions id must input"), 2)
	}

	// 查询是否有子集
	exists, err := l.svcCtx.DictionariesModel.ExistsWithParams(&orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: dictionaries.ColumnParentId, Values: functions.SliceToSliceAny(ids)},
		},
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	if exists {
		return nil, response.NewMakeRpcRetErr(errors.New("包含子集, 不能删除, 请先删除子集后操作"), 2)
	}

	return nil, response.NewMakeRpcRetErr(l.svcCtx.DictionariesModel.DeleteBy(params), 2)
}
