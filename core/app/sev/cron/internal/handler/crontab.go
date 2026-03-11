// @Title        crontab
// @Description  main
// @Create       yiyiyi 2025/7/11 15:44

package handler

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"skeyevss/core/app/sev/cron/internal/types"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/crontab"
)

type CrontabHandler struct {
	options []types.CrontabLogic

	svcCtx *types.ServiceContext
}

func NewCrontabHandler(svcCtx *types.ServiceContext) *CrontabHandler {
	return &CrontabHandler{
		svcCtx: svcCtx,
	}
}

func (h *CrontabHandler) Register(options ...types.CrontabLogic) {
	h.options = options

	go h.start()
	go h.setLogs()
}

func (h *CrontabHandler) start() {
	for v := range time.NewTicker(time.Second * 1).C {
		var now = v.Unix()
		for _, item := range h.options {
			record, ok := h.svcCtx.Data.Crontab[item.Key()]
			if !ok {
				continue
			}

			if record.Status == 0 || record.Interval == 0 {
				continue
			}

			if now%int64(record.Interval) != 0 {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(record.Timeout)*time.Second)
			defer cancel()

			go func() {
				if item.Executing() {
					return
				}

				item.DO(&types.CrontabLogicDOParams{
					Ctx:           ctx,
					SvcCtx:        h.svcCtx,
					Recover:       h.recover,
					CrontabRecord: record,
					Now:           now,
				})
			}()
		}
	}
}

func (h *CrontabHandler) setLogs() {
	time.Sleep(3 * time.Second)
	var data = &orm.BulkUpdateItem{
		Column: crontab.ColumnLogs,
	}
	for _, item := range h.svcCtx.Data.Crontab {
		b, err := functions.JSONMarshal(item.Logs)
		if err != nil {
			functions.LogError("定时任务logs序列化失败, err: ", err.Error, "; Id:", item.UniqueId)
			continue
		}

		data.Records = append(data.Records, &orm.BulkUpdateInner{
			PK:  item.UniqueId,
			Val: string(b),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := response.NewRpcToHttpResp[*deviceservice.Response, bool]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(h.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{BulkUpdates: []*orm.BulkUpdateItem{data}})
			if err != nil {
				return nil, err
			}

			return h.svcCtx.RpcClients.Config.CrontabBulkUpdate(ctx, data)
		},
	); err != nil {
		functions.LogError("定时任务数据logs更新失败, err: ", err.Error)
	}

	h.setLogs()
}

func (h *CrontabHandler) recover(name string) {
	if err := recover(); err != nil {
		functions.LogError(fmt.Sprintf("crontab [%s] Recover [%s] \nStack: %s", name, err, string(debug.Stack())))
	}
}
