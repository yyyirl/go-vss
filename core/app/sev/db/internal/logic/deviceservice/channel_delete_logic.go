package deviceservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/common/stream"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
)

type ChannelDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChannelDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChannelDeleteLogic {
	return &ChannelDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChannelDeleteLogic) ChannelDelete(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	// 获取通道
	list, err := l.svcCtx.ChannelsModel.List(&orm.ReqParams{
		Conditions: params.Conditions,
		All:        true,
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	// 删除通道
	if err := l.svcCtx.ChannelsModel.DeleteBy(params); err != nil {
		return nil, err
	}

	var streamNames []string
	for _, item := range list {
		streamNames = append(streamNames, stream.New().Produce(item.DeviceUniqueId, item.UniqueId, stream.PlayTypePlayback))
		streamNames = append(streamNames, stream.New().Produce(item.DeviceUniqueId, item.UniqueId, stream.PlayTypePlay))
	}

	return response.NewRpcResp[*db.Response]().Make(streamNames, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
