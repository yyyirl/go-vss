package deviceservicelogic

import (
	"context"
	"skeyevss/core/common/stream"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
)

type ChannelCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChannelCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChannelCreateLogic {
	return &ChannelCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChannelCreateLogic) ChannelCreate(in *db.MapReq) (*db.Response, error) {
	record, err := channels.NewItem().MapToModel(in.Data.AsMap())
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	if record.UniqueId == "" {
		record.UniqueId = functions.UniqueId()
	}

	data, err := record.ConvToModel(func(item *channels.Item) *channels.Item {
		return item
	})
	if err != nil || data == nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	data.Snapshot = stream.New().Snapshot(l.svcCtx.Config.SaveVideoSnapshotDir, data.DeviceUniqueId, data.UniqueId)

	res, err := l.svcCtx.ChannelsModel.Add(*data)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return &db.Response{
		Data:    []byte(strconv.Itoa(int(res.ID))),
		License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string),
	}, nil
}
