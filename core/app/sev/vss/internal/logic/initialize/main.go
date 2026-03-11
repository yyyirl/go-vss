// @Title        main
// @Description  main
// @Create       yiyiyi 2025/8/3 09:55

package initialize

import (
	"time"

	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/tps"
)

func DO(svcCtx *types.ServiceContext, baseConf *tps.YamlBaseConfig) {
	svcCtx.Broadcast.StartCleanupWorker(5*time.Second, 10*time.Second)

	// if err := redis.NewStreamKeepaliveRunningState(svcCtx.RedisClient).Clear(); err != nil {
	// 	functions.LogError("初始化清除保活状态失败, err:", err)
	// }
	//
	// if err := redis.NewQueue(svcCtx.RedisClient).Clear(redis.QueueStreamPubListen); err != nil {
	// 	functions.LogError("保活任务队列清除失败, err:", err)
	// }

	if svcCtx.Config.SipLogPath != "" {
		_ = functions.MakeDir(svcCtx.Config.SipLogPath)
	}

	// 更新sms默认配置
	// TODO 完整版请联系作者
}
