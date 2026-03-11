package deviceservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/devices"
)

type DeviceUpsertLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeviceUpsertLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeviceUpsertLogic {
	return &DeviceUpsertLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设备创建
func (l *DeviceUpsertLogic) DeviceUpsert(in *db.MapReq) (*db.Response, error) {
	record, err := devices.NewItem().MapToModel(in.Data.AsMap())
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	data, err := record.ConvToModel(nil)
	if err != nil || data == nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	var neededUpdateColumns = []string{
		devices.ColumnName,
		devices.ColumnAccessProtocol,
		devices.ColumnDeviceUniqueId,
		devices.ColumnState,
		devices.ColumnOnline,
		devices.ColumnExpire,
		devices.ColumnAddress,
		devices.ColumnRegisterAt,
	}
	return nil, response.NewMakeRpcRetErr(
		l.svcCtx.DevicesModel.UpsertWithExcludeColumns(
			[]devices.Devices{
				{
					Name:           data.Name,
					AccessProtocol: data.AccessProtocol,
					DeviceUniqueId: data.DeviceUniqueId,
					State:          data.State,
					Online:         data.Online,
					Expire:         data.Expire,
					Address:        data.Address,
					RegisterAt:     data.RegisterAt,
				},
			},
			[]string{devices.ColumnDeviceUniqueId},
			functions.ArrFilter(devices.Columns, func(item string) bool {
				return !functions.Contains(item, neededUpdateColumns)
			}),
		),
		2,
	)
}
