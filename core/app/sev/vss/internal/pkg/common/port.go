// @Title        port
// @Description  main
// @Create       yiyiyi 2025/12/30 21:13

package common

import (
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/functions/sc"
)

func UsablePort(svcCtx *types.ServiceContext) uint {
	// 检测请求合法性
	var (
		all   = svcCtx.TalkSipData.All()
		ports []int
	)
	for _, item := range all {
		if item.RTPUsablePort > 0 {
			ports = append(ports, item.RTPUsablePort)
		}
	}

	// 本地端口
	var usablePort uint
	for i := svcCtx.Config.Sip.UsableMinPort; i <= svcCtx.Config.Sip.UsableMaxPort; i++ {
		if functions.Contains(int(i), ports) {
			continue
		}

		state, _ := sc.CheckPort(int(i), 0)
		if !state {
			usablePort = i
			break
		}
	}

	return usablePort
}
