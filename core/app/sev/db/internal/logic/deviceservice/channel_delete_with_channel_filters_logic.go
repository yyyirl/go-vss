package deviceservicelogic

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
)

type ChannelDeleteWithChannelFiltersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChannelDeleteWithChannelFiltersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChannelDeleteWithChannelFiltersLogic {
	return &ChannelDeleteWithChannelFiltersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除通道(Devices.ChannelFilters)
func (l *ChannelDeleteWithChannelFiltersLogic) ChannelDeleteWithChannelFilters(in *db.UniqueIdsReq) (*db.Response, error) {
	if err := l.svcCtx.ChannelsModel.DeleteWithChannelFilters(l.ctx, in.UniqueId, in.UniqueIds); err != nil {
		return nil, err
	}

	return &db.Response{Data: []byte(strconv.FormatBool(true))}, nil
}
