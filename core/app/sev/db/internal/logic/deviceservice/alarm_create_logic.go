package deviceservicelogic

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/alarms"
)

type AlarmCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAlarmCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AlarmCreateLogic {
	return &AlarmCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AlarmCreateLogic) AlarmCreate(in *db.MapReq) (*db.Response, error) {
	record, err := alarms.NewItem().MapToModel(in.Data.AsMap())
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	data, err := record.ConvToModel(func(item *alarms.Item) *alarms.Item {
		return item
	})
	if err != nil || data == nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	res, err := l.svcCtx.AlarmsModel.Add(*data)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return &db.Response{
		Data:    []byte(strconv.Itoa(int(res.ID))),
		License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string),
	}, nil
}
