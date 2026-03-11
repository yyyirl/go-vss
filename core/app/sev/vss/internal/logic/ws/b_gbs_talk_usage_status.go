package ws

import (
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
)

const BGBSTalkUsageStatusKey = "gbs-talk-usage-status-broadcast"

type BGBSTalkUsageStatusLogic struct {
	*broadcasts
	svcCtx *types.ServiceContext
}

func NewBGBSTalkUsageStatus(svcCtx *types.ServiceContext) *BGBSTalkUsageStatusLogic {
	return &BGBSTalkUsageStatusLogic{svcCtx: svcCtx, broadcasts: &broadcasts{svcCtx: svcCtx}}
}

func (l *BGBSTalkUsageStatusLogic) Do(req *types.BroadcastMessageTalkUsageStatus) error {
	l.sendWithActivateKey(BGBSTalkUsageStatusKey, req.Key, req)
	return nil
}

func BGBSSendTalkUsageStatus(svcCtx *types.ServiceContext, uniqueId, key string, state uint) {
	svcCtx.WSProc.BroadcastChan <- &types.BroadcastMessageItem{
		Type:   BGBSTalkUsageStatusKey,
		Caller: functions.Caller(2),
		Data: &types.BroadcastMessageTalkUsageStatus{
			Key:      key,
			State:    state,
			UniqueId: uniqueId,
		},
	}
}
