package env

import (
	"context"
	"fmt"

	xconf "github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/common"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
	"skeyevss/core/tps/conf"
)

type UpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateLogic) Update(req *types.SetEnvReq) *response.HttpErr {
	if req.Content == "" {
		return response.MakeError(response.NewHttpRespMessage().Str("content 不能为空"), localization.M0001)
	}

	var envFileBackup = fmt.Sprintf("%s.tmp.%s", l.svcCtx.Config.EnvFile, functions.NewTimer().Format(functions.TimeFormatYmdhis))
	// 备份.env
	if err := functions.WriteToFile(envFileBackup, req.Content); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00400)
	}

	defer functions.OverloadEnvFile(l.svcCtx.Config.EnvFile)
	// 检测env合法性
	functions.OverloadEnvFile(envFileBackup)

	var backendApiConf conf.BackendApiConf
	if err := xconf.Load(l.svcCtx.Config.ConfigPath.BackendApiConf, &backendApiConf, xconf.UseEnv()); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00401)
	}

	var cronConfig conf.CronConfig
	if err := xconf.Load(l.svcCtx.Config.ConfigPath.CronConf, &cronConfig, xconf.UseEnv()); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00401)
	}

	var dbSevConf conf.DBSevConf
	if err := xconf.Load(l.svcCtx.Config.ConfigPath.DbGrpcConf, &dbSevConf, xconf.UseEnv()); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00401)
	}

	var vssSevConfig conf.VssSevConfig
	if err := xconf.Load(l.svcCtx.Config.ConfigPath.VssConf, &vssSevConfig, xconf.UseEnv()); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00401)
	}

	var webConfig conf.WebConfig
	if err := xconf.Load(l.svcCtx.Config.ConfigPath.WebSevConf, &webConfig, xconf.UseEnv()); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00401)
	}

	// 替换env
	if err := functions.Mv(envFileBackup, l.svcCtx.Config.EnvFile); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00402)
	}

	return common.Restart(l.svcCtx, "")
}
