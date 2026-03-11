// @Title        data
// @Description  main
// @Create       yiyiyi 2025/7/8 13:54

package handler

import (
	"sync"
	"time"

	"skeyevss/core/app/sev/cron/internal/types"
)

var (
	fetchDataWg    sync.WaitGroup
	fetchDataIsSet = false
)

type FetchDataLogic struct {
	svcCtx *types.ServiceContext
}

func NewFetchDataLogic(svcCtx *types.ServiceContext) *FetchDataLogic {
	return &FetchDataLogic{
		svcCtx: svcCtx,
	}
}

func (h *FetchDataLogic) DO(options ...types.FetchDataProcLogic) {
	fetchDataWg.Add(len(options))
	h.do(options...)
	fetchDataWg.Wait()
	fetchDataIsSet = true

	for range time.NewTicker(time.Second * 3).C {
		h.do(options...)
	}
}

func (h *FetchDataLogic) do(options ...types.FetchDataProcLogic) {
	for _, item := range options {
		go func() {
			// 第一次获取 未进入定时器
			if !fetchDataIsSet {
				defer fetchDataWg.Done()
			}

			item.DO(&types.FetchDataLogicParams{SvcCtx: h.svcCtx})
		}()
	}
}
