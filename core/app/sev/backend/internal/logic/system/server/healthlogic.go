package server

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/pkg/device"
	"skeyevss/core/pkg/functions"
)

var diskUsage []device.DiskUsageInfo

type HealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HealthLogic) Health() interface{} {
	var (
		hardware = l.svcCtx.Health.Hardware
		services = l.svcCtx.Health.Services
	)
	if len(hardware) > l.svcCtx.Health.Count {
		hardware = hardware[len(hardware)-l.svcCtx.Health.Count:]
	}

	if len(services) > l.svcCtx.Health.Count {
		services = services[len(services)-l.svcCtx.Health.Count:]
	}

	var now = functions.NewTimer().Now()
	if len(diskUsage) <= 0 || now%20 == 0 {
		var err error
		diskUsage, err = device.NewSystem().GetDiskUsage()
		if err != nil {
			functions.LogcError(l.ctx, "磁盘信息获取失败, err: ", err.Error())
		}
	}

	return map[string]interface{}{
		"hardware":         hardware,
		"services":         services,
		"memTotal":         l.svcCtx.Health.MemTotal,
		"diskUsage":        diskUsage,
		"deviceStatistics": l.svcCtx.DeviceStatistics(),
	}
}
