package configservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/dictionaries"
)

type DictionaryUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDictionaryUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictionaryUpdateLogic {
	return &DictionaryUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DictionaryUpdateLogic) DictionaryUpdate(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	record, err := dictionaries.NewItem().CheckMap(params.DataRecord)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return nil, response.NewMakeRpcRetErr(l.svcCtx.DictionariesModel.UpdateWithParams(record, params), 2)
}
