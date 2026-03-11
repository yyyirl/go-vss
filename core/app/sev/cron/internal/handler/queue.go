// @Title        queue
// @Description  main
// @Create       yiyiyi 2025/7/8 10:06

package handler

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"skeyevss/core/app/sev/cron/internal/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/repositories/redis"
)

type QueueHandler struct {
	options []types.QueueLogic

	svcCtx *types.ServiceContext
}

func NewQueueHandler(svcCtx *types.ServiceContext) *QueueHandler {
	return &QueueHandler{
		svcCtx: svcCtx,
	}
}

func (h *QueueHandler) Register(options ...types.QueueLogic) {
	h.options = options

	var maps = make(map[string]bool)
	for _, item := range h.options {
		if _, ok := maps[item.Key()]; ok {
			panic(fmt.Sprintf("duplicate queue key: %s", item.Key()))
		}

		maps[item.Key()] = true
	}

	go h.start()
}

func (h *QueueHandler) start() {
	for range time.NewTicker(time.Second * 1).C {
		for _, item := range h.options {
			go func() {
				if item.Executing() {
					return
				}

				defer func() {
					item.SetExecuting(false)
				}()

				item.SetExecuting(true)

				ctx, cancel := context.WithTimeout(context.Background(), item.Timeout())
				defer cancel()

				if err := redis.NewQueue(h.svcCtx.RedisClient).Get(item.Key(), item.Limit(), func(data [][]byte) {
					if err := item.DO(&types.QueueLogicDOParams{
						Ctx:     ctx,
						SvcCtx:  h.svcCtx,
						Recover: h.recover,
						Data:    data,
					}); err != nil {
						functions.LogError("消息队列[", item.Key(), "]执行失败, err: ", err)
					}
				}); err != nil {
					functions.LogError("消息队列[", item.Key(), "]数据获取失败, err: ", err)
				}
			}()
		}
	}
}

func (h *QueueHandler) recover(name string) {
	if err := recover(); err != nil {
		functions.LogError(fmt.Sprintf("queue [%s] Recover [%s] \nStack: %s", name, err, string(debug.Stack())))
	}
}
