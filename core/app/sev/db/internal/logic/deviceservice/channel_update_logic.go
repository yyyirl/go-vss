package deviceservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
)

type ChannelUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChannelUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChannelUpdateLogic {
	return &ChannelUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChannelUpdateLogic) ChannelUpdate(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	if _, ok := params.DataRecord[channels.ColumnDepIds]; ok {
		// 设置设备depId更新
		l.svcCtx.DeviceDepIdSetChan <- struct{}{}
	}

	record, err := channels.NewItem().CheckMap(params.DataRecord)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return nil, response.NewMakeRpcRetErr(l.svcCtx.ChannelsModel.UpdateWithParams(record, params), 2)
}
