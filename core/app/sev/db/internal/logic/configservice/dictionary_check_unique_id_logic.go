package configservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/dictionaries"
)

type DictionaryCheckUniqueIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDictionaryCheckUniqueIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictionaryCheckUniqueIdLogic {
	return &DictionaryCheckUniqueIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DictionaryCheckUniqueIdLogic) DictionaryCheckUniqueId(in *db.UniqueIdReq) (*db.Response, error) {
	exists, err := l.svcCtx.DictionariesModel.ExistsWithParams(&orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: dictionaries.ColumnUniqueId, Value: in.UniqueId},
		},
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return &db.Response{
		Data:    functions.BoolToByte(exists),
		License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string),
	}, nil
}
