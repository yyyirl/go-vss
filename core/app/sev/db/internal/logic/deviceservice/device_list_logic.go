package deviceservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/devices"
)

type DeviceListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeviceListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeviceListLogic {
	return &DeviceListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeviceListLogic) DeviceList(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	// 获取总数
	count, queryErr := l.svcCtx.DevicesModel.Count(params)
	if queryErr != nil {
		return nil, response.NewMakeRpcRetErr(queryErr, 2)
	}

	if count <= 0 {
		return response.NewRpcResp[*db.Response]().Make(response.NewListResp[[]*devices.Item]().Empty(), 3, func(data []byte) *db.Response {
			return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
		})
	}

	// 获取列表
	list, queryErr := l.svcCtx.DevicesModel.List(params)
	if queryErr != nil {
		return nil, response.NewMakeRpcRetErr(queryErr, 2)
	}

	var records []*devices.Item
	for _, item := range list {
		v, err := item.ConvToItem()
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		records = append(records, v)
	}

	// 获取列表
	onvifList, _ := l.svcCtx.DevicesModel.List(&orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: devices.ColumnAccessProtocol, Value: devices.AccessProtocol_3},
			{Column: devices.ColumnAddress, Value: "", Operator: "!="},
		},
	})
	var onvifAddresses []string
	if len(onvifList) > 0 {
		for _, item := range onvifList {
			onvifAddresses = append(onvifAddresses, item.Address)
		}
	}

	return response.NewRpcResp[*db.Response]().Make(&response.ListWithExtResp[[]*devices.Item, []string, string]{
		List:   records,
		Count:  count,
		Slices: onvifAddresses,
	}, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
