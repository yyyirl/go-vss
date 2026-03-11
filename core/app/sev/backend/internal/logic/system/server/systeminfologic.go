package server

import (
	"context"
	"fmt"
	"runtime"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/pkg/functions"
)

type SystemInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSystemInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SystemInfoLogic {
	return &SystemInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SystemInfoLogic) SystemInfo() map[string]interface{} {
	return map[string]interface{}{
		// 服务启动时间
		"sevStartTime": l.svcCtx.StartTimestamp,
		// 服务器时间
		"sevTime": functions.NewTimer().NowMilli(),
		// 构建信息
		"buildVersion": fmt.Sprintf(
			"%s/%s (platfrom/%s; buildAt/%s)",
			l.svcCtx.Config.ProductName,
			l.svcCtx.Config.Version,
			runtime.GOOS,
			l.svcCtx.BuildTime,
		),
		"channel":       0,
		"channelOnline": l.svcCtx.DeviceStatistics().ChannelOnlineCount,
		"osEnvironment": l.svcCtx.Config.OSEnvironment,
	}
}
