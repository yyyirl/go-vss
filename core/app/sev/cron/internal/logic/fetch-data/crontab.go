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
	"skeyevss/core/repositories/models/crontab"
)

var _ types.FetchDataProcLogic = (*FetchCrontabLogic)(nil)

var fetchCrontab sync.Once

type FetchCrontabLogic struct {
}

func (l *FetchCrontabLogic) DO(params *types.FetchDataLogicParams) {
	l.do(params)
}

func (l *FetchCrontabLogic) do(params *types.FetchDataLogicParams) {
	defer fetchCrontab.Do(func() {
		go params.SvcCtx.Data.InitFetchDataState.Done()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := response.NewRpcToHttpResp[*configservice.Response, []*crontab.Item]().Parse(
		func() (*configservice.Response, error) {
			return params.SvcCtx.RpcClients.Config.Crontab(ctx, &configservice.EmptyRequest{})
		},
	)
	if err != nil {
		functions.LogError("do fetch crontab fail", err.Error)
		return
	}

	var maps = make(map[string]*crontab.Item)
	for _, item := range res.Data {
		maps[item.UniqueId] = item
	}

	params.SvcCtx.Data.Crontab = maps
}
