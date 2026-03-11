// @Title        main
// @Description  main
// @Create       yiyiyi 2025/7/8 10:06

package types

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"skeyevss/core/app/sev/cron/internal/config"
	"skeyevss/core/common/client"
	cTypes "skeyevss/core/common/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/repositories/models/crontab"
	mediaServers "skeyevss/core/repositories/models/media-servers"
	"skeyevss/core/repositories/models/settings"
	"skeyevss/core/repositories/redis"
)

type (
	SvcData struct {
		Setting      *settings.Item                // 设置
		Crontab      map[string]*crontab.Item      // 任务列表
		MediaServers map[uint64]*mediaServers.Item // 媒体服务

		// 初始化数据加载完成状态
		InitFetchDataState sync.WaitGroup
	}

	ServiceContext struct {
		Config      config.Config
		RedisClient *redis.Client
		RpcClients  *client.GRPCClients

		Data *SvcData
	}

	QueueLogicDOParams struct {
		Ctx     context.Context
		SvcCtx  *ServiceContext
		Recover func(name string)
		Data    [][]byte
	}

	QueueLogic interface {
		DO(params *QueueLogicDOParams) error // execute
		Timeout() time.Duration              // 执行超时时间
		Limit() int                          // 队列每一批次取出数量
		Executing() bool                     // 当前批次是否执行完成
		SetExecuting(v bool)                 // 设置完成
		Key() string                         // 队列cache key
	}

	CrontabLogicDOParams struct {
		Ctx           context.Context
		SvcCtx        *ServiceContext
		Recover       func(name string)
		CrontabRecord *crontab.Item
		Now           int64
	}

	CrontabLogic interface {
		DO(params *CrontabLogicDOParams) // execute
		Executing() bool                 // 当前批次是否正在执行
		Key() string                     // 任务key
	}

	// interval
	FetchDataLogicParams struct {
		SvcCtx *ServiceContext
	}

	FetchDataProcLogic interface {
		DO(params *FetchDataLogicParams)
	}
)

func (s *ServiceContext) Settings() *settings.Item {
	s.Data.Setting.ItemCorrection(&settings.ItemCorrectionParams{
		BaseConf:   s.Config.SevBase,
		SipConf:    s.Config.Sip,
		InternalIp: s.Config.InternalIP,
		ExternalIp: s.Config.ExternalIP,
	})

	return s.Data.Setting
}

func (s *ServiceContext) MSVoteNode(ids []uint64) *cTypes.MSVoteNodeResp {
	// 默认节点
	var (
		mediaServerInternalIP   = s.Config.InternalIP
		mediaServerInternalPort = s.Config.SevBase.MediaServerPort

		mediaServerExternalIP   = s.Config.ExternalIP
		mediaServerExternalPort = s.Config.SevBase.MediaServerPort

		node string
	)
	if mediaServerInternalIP != "" && mediaServerInternalPort > 0 {
		node = fmt.Sprintf("%s:%d", mediaServerInternalIP, mediaServerInternalPort)
	} else if mediaServerExternalIP != "" && mediaServerExternalPort > 0 {
		node = fmt.Sprintf("%s:%d", mediaServerExternalIP, mediaServerExternalPort)
	}

	if len(ids) <= 0 {
		return &cTypes.MSVoteNodeResp{
			Address: node,
			Name:    "default",
		}
	}

	var nodes []*cTypes.MSVoteNodeResp
	for _, item := range s.Data.MediaServers {
		if functions.Contains(item.ID, ids) && item.IP != "" && item.Port > 0 {
			nodes = append(nodes, &cTypes.MSVoteNodeResp{
				Address: fmt.Sprintf("%s:%d", item.IP, item.Port),
				Name:    item.Name,
				ID:      item.ID,
			})
		}
	}

	if len(nodes) <= 0 {
		return &cTypes.MSVoteNodeResp{
			Address: node,
			Name:    "default",
		}
	}

	if len(nodes) <= 1 {
		return &cTypes.MSVoteNodeResp{
			Address: node,
			Name:    "default",
		}
	}

	return nodes[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(nodes))]
}
