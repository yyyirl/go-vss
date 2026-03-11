package gbs_proc

import (
	"skeyevss/core/app/sev/vss/internal/types"
)

var _ types.SipProcLogic = (*SipLogLogic)(nil)

type SipLogLogic struct {
	svcCtx      *types.ServiceContext
	recoverCall func(name string)
}

func (l *SipLogLogic) DO(params *types.DOProcLogicParams) {
	l = &SipLogLogic{
		svcCtx:      params.SvcCtx,
		recoverCall: params.RecoverCall,
	}
	l.svcCtx.InitFetchDataState.Wait()

	defer l.recoverCall("siplog接收")

	for {
		select {
		case v := <-l.svcCtx.SipLog:
			l.svcCtx.Broadcast.Send(v.Type, v.Content)
		}
	}
}
