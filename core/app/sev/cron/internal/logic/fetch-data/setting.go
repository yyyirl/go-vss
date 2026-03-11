// @Title        setting
// @Description  main
// @Create       yiyiyi 2025/7/8 13:50

package fetchdata

import (
	"context"
	"sync"
	"time"

	"skeyevss/core/app/sev/cron/internal/types"
	"skeyevss/core/app/sev/db/client/configservice"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/settings"
)

var _ types.FetchDataProcLogic = (*FetchSettingLogic)(nil)

var fetchSetting sync.Once

type FetchSettingLogic struct {
}

func (l *FetchSettingLogic) DO(params *types.FetchDataLogicParams) {
	l.do(params)
}

func (l *FetchSettingLogic) do(params *types.FetchDataLogicParams) {
	defer fetchSetting.Do(func() {
		go params.SvcCtx.Data.InitFetchDataState.Done()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := response.NewRpcToHttpResp[*configservice.Response, *settings.Item]().Parse(
		func() (*configservice.Response, error) {
			return params.SvcCtx.RpcClients.Config.SettingRow(ctx, &configservice.EmptyRequest{})
		},
	)
	if err != nil {
		functions.LogError("do fetch crontab fail", err.Error)
		return
	}

	params.SvcCtx.Data.Setting = res.Data
}
