package deviceservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/common/stream"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
)

type ChannelUpsertLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChannelUpsertLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChannelUpsertLogic {
	return &ChannelUpsertLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChannelUpsertLogic) ChannelUpsert(in *db.SliceMapReq) (*db.Response, error) {
	var (
		records, records1   []channels.Channels
		neededUpdateColumns = []string{
			channels.ColumnName,
			channels.ColumnUniqueId,
			channels.ColumnDeviceUniqueId,
			channels.ColumnOnline,
			channels.ColumnParentID,
			channels.ColumnParental,
			channels.ColumnOnlineAt,
			channels.ColumnIsCascade,
			channels.ColumnStreamUrl,
			channels.ColumnPtzType,
			channels.ColumnOriginalChannelUniqueId,
			channels.ColumnOriginal,
			channels.ColumnSnapshot,
		}
		onConflictColumns = []string{
			channels.ColumnUniqueId,
			channels.ColumnDeviceUniqueId,
		}
	)
	for _, item := range in.Data {
		var model channels.Channels
		if err := functions.ConvInterface(item.AsMap(), &model); err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		model.Snapshot = stream.New().Snapshot(l.svcCtx.Config.SaveVideoSnapshotDir, model.DeviceUniqueId, model.UniqueId)
		if model.Longitude <= 0 || model.Latitude <= 0 {
			records1 = append(records1, model)
			continue
		}

		records = append(records, model)
	}

	if len(records) > 0 {
		return nil, response.NewMakeRpcRetErr(
			l.svcCtx.ChannelsModel.UpsertWithExcludeColumns(
				records,
				onConflictColumns,
				functions.ArrFilter(channels.Columns, func(item string) bool {
					return !functions.Contains(
						item,
						append(
							neededUpdateColumns,
							channels.ColumnLongitude,
							channels.ColumnLatitude,
						),
					)
				}),
			),
			2,
		)
	}

	// 如果数据库有已经设置的经纬度值且catalog传递的经纬度为空则忽略更新
	if len(records1) > 0 {
		return nil, response.NewMakeRpcRetErr(
			l.svcCtx.ChannelsModel.UpsertWithExcludeColumns(
				records1,
				onConflictColumns,
				functions.ArrFilter(channels.Columns, func(item string) bool {
					return !functions.Contains(item, neededUpdateColumns)
				}),
			),
			2,
		)
	}

	return nil, nil
}
