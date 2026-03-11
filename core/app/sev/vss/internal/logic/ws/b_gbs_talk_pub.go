package ws

import (
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
)

const BGBSTalkPubKey = "gbs-talk-pub-broadcast"

type BGBSTalkPubLogic struct {
	*broadcasts
	svcCtx *types.ServiceContext
}

func NewBGBSTalkPub(svcCtx *types.ServiceContext) *BGBSTalkPubLogic {
	return &BGBSTalkPubLogic{svcCtx: svcCtx, broadcasts: &broadcasts{svcCtx: svcCtx}}
}

func (l *BGBSTalkPubLogic) Do(req *types.BroadcastMessageTalkSipState) error {
	l.sendWithActivateKey(BGBSTalkPubKey, req.Key, req)
	return nil
}

func BGBSSendTalkPub(svcCtx *types.ServiceContext, key string, state uint) {
	svcCtx.WSProc.BroadcastChan <- &types.BroadcastMessageItem{
		Type:   BGBSTalkPubKey,
		Caller: functions.Caller(2),
		Data: &types.BroadcastMessageTalkSipState{
			Key:   key,
			State: state,
		},
	}
}

func BGBSSendTalkPubError(svcCtx *types.ServiceContext, key string, msg string) {
	svcCtx.WSProc.BroadcastChan <- &types.BroadcastMessageItem{
		Type:   BGBSTalkPubKey,
		Caller: functions.Caller(2),
		Data: &types.BroadcastMessageTalkSipState{
			Key:          key,
			FailedReason: msg,
		},
	}
}
