// @Title        获取流媒体服务记录
// @Description  main
// @Create       yiyiyi 2025/7/8 13:50

package fetchdata

import (
	"context"
	"sync"
	"time"

	"skeyevss/core/app/sev/cron/internal/types"
	"skeyevss/core/app/sev/db/client/configservice"
	configClient "skeyevss/core/app/sev/db/client/configservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	mediaServers "skeyevss/core/repositories/models/media-servers"
)

var _ types.FetchDataProcLogic = (*FetchMediaServerLogic)(nil)

var fetchMediaServer sync.Once

type FetchMediaServerLogic struct {
}

func (l *FetchMediaServerLogic) DO(params *types.FetchDataLogicParams) {
	l.do(params)
}

func (l *FetchMediaServerLogic) do(params *types.FetchDataLogicParams) {
	defer fetchMediaServer.Do(func() {
		go params.SvcCtx.Data.InitFetchDataState.Done()
	})

	// 获取资源
	res, err := response.NewRpcToHttpResp[*configservice.Response, *response.ListResp[[]*mediaServers.Item]]().Parse(
		func() (*configClient.Response, error) {
			data, err := conv.New(params.SvcCtx.Config.Mode).ToPBParams(&orm.ReqParams{All: true})
			if err != nil {
				return nil, err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			return params.SvcCtx.RpcClients.Config.Dictionaries(ctx, data)
		},
	)
	if err != nil {
		functions.LogError("fetch media server crontab fail", err.Error)
		return
	}

	var maps = make(map[uint64]*mediaServers.Item)
	for _, item := range res.Data.List {
		maps[item.ID] = item
	}

	params.SvcCtx.Data.MediaServers = maps
}
