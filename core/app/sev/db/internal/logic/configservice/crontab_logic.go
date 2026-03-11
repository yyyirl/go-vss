package configservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/crontab"
)

type CrontabLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCrontabLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CrontabLogic {
	return &CrontabLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CrontabLogic) Crontab(_ *db.EmptyRequest) (*db.Response, error) {
	// 获取列表
	list, err := l.svcCtx.CrontabModel.List(&orm.ReqParams{All: true})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	var records []*crontab.Item
	for _, item := range list {
		v, err := item.ConvToItem()
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		records = append(records, v)
	}

	return response.NewRpcResp[*db.Response]().Make(records, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
