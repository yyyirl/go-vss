package gbs_proc

import (
	"time"

	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
)

var _ types.SipProcLogic = (*CatalogLoopLogic)(nil)

type CatalogLoopLogic struct {
	svcCtx      *types.ServiceContext
	recoverCall func(name string)
}

// 定时发送catalog
func (l *CatalogLoopLogic) DO(params *types.DOProcLogicParams) {
	l = &CatalogLoopLogic{
		svcCtx:      params.SvcCtx,
		recoverCall: params.RecoverCall,
	}
	l.svcCtx.InitFetchDataState.Wait()

	defer l.recoverCall("定时发送catalog")

	// 创建定时器
	go l.proc()

	for {
		select {
		case v := <-l.svcCtx.SipCatalogLoop:
			l.make(v)
		}
	}
}

func (l *CatalogLoopLogic) make(data *types.SipCatalogLoopReq) {
	if data.Online {
		// 注册定时器
		l.svcCtx.SipCatalogLoopMap.Set(data.Req.ID, data)
	} else {
		// 删除定时器
		l.svcCtx.SipCatalogLoopMap.Remove(data.Req.ID)
	}
}

func (l *CatalogLoopLogic) proc() {
	defer l.recoverCall("定时发送catalog loop")

	for val := range time.NewTicker(time.Second * 1).C {
		// functions.LogInfo("定时器 catalog_loop l.svcCtx.SipCatalogLoopMap.Values: ", l.svcCtx.SipCatalogLoopMap.Len())
		for _, item := range l.svcCtx.SipCatalogLoopMap.Values() {
			if item.Now%val.Unix() != l.svcCtx.Config.Sip.CatalogInterval || !item.Online {
				continue
			}

			item.Req.Caller = functions.CallerFile(1)
			// 发送catalog
			l.svcCtx.SipSendCatalog <- item.Req
		}
	}
}
