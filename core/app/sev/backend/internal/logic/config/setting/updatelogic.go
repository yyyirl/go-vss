package setting

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/configservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/common/opt"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/settings"
	systemOperationLogs "skeyevss/core/repositories/models/system-operation-logs"
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

func (l *UpdateLogic) Update(req *orm.ReqParams) *response.HttpErr {
	// 日志记录
	opt.NewSystemOperationLogs(l.svcCtx.RpcClients).Make(l.ctx, systemOperationLogs.Types[systemOperationLogs.TypeSettingUpdate], req)

	var content settings.Content
	for _, item := range req.Data {
		if item.Column == settings.ColumnContent {
			if err := functions.ConvInterface(item.Value, &content); err != nil {
				return response.MakeError(response.NewHttpRespMessage().Err(err), localization.MR1002)
			}
		}
	}

	var settingContent = l.svcCtx.Settings().Content
	if settingContent != nil {
		if settingContent.MapZoom <= 6 || settingContent.MapZoom > 12 {
			settingContent.MapZoom = 6
		}
	}

	// 解压瓦片图
	if content.MapTiles != "" && settingContent == nil || (settingContent != nil && content.MapTiles != settingContent.MapTiles) {
		var tileBaseDir = "source/maps/" + functions.NewTimer().Format(functions.TimeFormatYmdhis)
		if err := functions.UnZip(content.MapTiles, tileBaseDir); err != nil {
			return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00403)
		}

		tilesDir, err := l.findFirstTilesFolder(tileBaseDir)
		if err != nil {
			return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00403)
		}

		content.MapTiles = fmt.Sprintf("%s/%s", tileBaseDir, tilesDir)
		for _, item := range req.Data {
			if item.Column == settings.ColumnContent {
				item.Value = functions.StructToMap(content, "json", nil)
			}
		}
	}

	_, err := response.NewRpcToHttpResp[*configservice.Response, bool]().Parse(
		func() (*configservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(req)
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Config.SettingUpdate(l.ctx, data)
		},
	)
	if err == nil {
		l.svcCtx.SettingSet <- struct{}{}
	}

	return err
}

func (l *UpdateLogic) findFirstTilesFolder(dirPath string) (string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			folderName := entry.Name()
			if strings.HasSuffix(folderName, "_tiles") {
				return folderName, nil
			}
		}
	}

	return "", fmt.Errorf("未找到符合条件的文件夹")
}
