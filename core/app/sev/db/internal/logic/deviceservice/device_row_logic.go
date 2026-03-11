package deviceservicelogic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
)

type DeviceRowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeviceRowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeviceRowLogic {
	return &DeviceRowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeviceRowLogic) DeviceRow(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	row, err := l.svcCtx.DevicesModel.RowWithParams(params)
	if err != nil {
		if params.IgnoreNotFound && errors.Is(err, orm.NotFound) {
			return response.NewRpcResp[*db.Response]().Make(nil, 3, func(data []byte) *db.Response {
				return &db.Response{Data: nil, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
			})
		}

		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	data, err := row.ConvToItem()
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return response.NewRpcResp[*db.Response]().Make(data, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
