package devices

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/common/stream"
	cTypes "skeyevss/core/common/types"
	"skeyevss/core/common/videoProject"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
)

type DownloadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadLogic {
	return &DownloadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadLogic) Download(req *types.VideoDeviceDownloadReq) (string, *response.HttpErr) {
	// 获取通道
	channelMaps, err := response.NewRpcToHttpResp[*deviceservice.Response, map[uint64]*cTypes.ChannelMSRelItem]().Parse(
		func() (*backendservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: channels.ColumnUniqueId, Value: req.ChannelUniqueId},
				},
				All: true,
			})
			if err != nil {
				return nil, response.NewMakeRpcRetErr(err, 2)
			}

			return l.svcCtx.RpcClients.Device.MediaServersWithChannelIds(l.ctx, data)
		},
	)
	if err != nil {
		return "", err
	}

	var channelId uint64
	for _, item := range channelMaps.Data {
		if item.ChannelUniqueId == req.ChannelUniqueId {
			channelId = item.ChannelId
			break
		}
	}

	if channelId <= 0 {
		return "", response.MakeError(response.NewHttpRespMessage().Str("通道获取失败"), localization.M0010)
	}

	var (
		streamName = stream.New().ProduceWith(
			channelMaps.Data[channelId].DeviceUniqueId,
			channelMaps.Data[channelId].ChannelUniqueId,
			stream.PlayTypePlayback,
			req.UniqueId,
		)
		rq = l.svcCtx.RemoteReq(l.ctx)
	)
	videoProject.NewRecoding().StartRecording(
		&videoProject.Params{
			IsDownloading: true,
			VssHttpTarget: rq.VssHttpTarget,
			Mode:          l.svcCtx.Config.Mode,
			GetMSAddress:  func(msIds []uint64) string { return l.svcCtx.MSVoteNode(msIds).Address },
			RpcClients:    l.svcCtx.RpcClients,
			PlayType:      stream.PlayTypePlayback,
			StartAt:       req.StartAt,
			EndAt:         req.EndAt,
			StreamName:    streamName,
		},
		map[uint64]*cTypes.ChannelMSRelItem{
			channelId: channelMaps.Data[channelId],
		},
	)

	return streamName, nil
}
