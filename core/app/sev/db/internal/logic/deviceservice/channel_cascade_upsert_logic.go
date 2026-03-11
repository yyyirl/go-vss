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

type ChannelCascadeUpsertLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChannelCascadeUpsertLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChannelCascadeUpsertLogic {
	return &ChannelCascadeUpsertLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 通道upsert
func (l *ChannelCascadeUpsertLogic) ChannelCascadeUpsert(in *db.SliceMapReq) (*db.Response, error) {
	var records []channels.Channels
	for _, item := range in.Data {
		var model channels.Channels
		if err := functions.ConvInterface(item.AsMap(), &model); err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		model.Snapshot = stream.New().Snapshot(l.svcCtx.Config.SaveVideoSnapshotDir, model.DeviceUniqueId, model.UniqueId)
		records = append(records, model)
	}

	var neededUpdateColumns = []string{
		channels.ColumnID,
		channels.ColumnCascadeChannelUniqueId,
		channels.ColumnUniqueId,
		channels.ColumnDeviceUniqueId,
	}
	return nil, response.NewMakeRpcRetErr(
		l.svcCtx.ChannelsModel.UpsertWithExcludeColumns(
			records,
			[]string{
				channels.ColumnID,
				channels.ColumnCascadeChannelUniqueId,
				channels.ColumnUniqueId,
			},
			functions.ArrFilter(channels.Columns, func(item string) bool {
				return !functions.Contains(item, neededUpdateColumns)
			}),
		),
		2,
	)
}
