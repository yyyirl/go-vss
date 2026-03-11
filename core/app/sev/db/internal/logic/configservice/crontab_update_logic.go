package configservicelogic

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/crontab"
)

type CrontabUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCrontabUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CrontabUpdateLogic {
	return &CrontabUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CrontabUpdateLogic) CrontabUpdate(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	if logs, ok := params.DataRecord[crontab.ColumnLogs]; ok {
		tmp, ok := logs.([]interface{})
		if !ok {
			return nil, response.NewMakeRpcRetErr(fmt.Errorf("logs type err input: %T need []string", logs), 2)
		}

		var logs []string
		for _, v := range tmp {
			record, ok := v.(string)
			if !ok {
				return nil, response.NewMakeRpcRetErr(fmt.Errorf("logs item type err input: %T need string", logs), 2)
			}

			logs = append(logs, record)
		}

		row, err := l.svcCtx.CrontabModel.RowWithParams(params)
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		item, err := row.ConvToItem()
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		logs = append(logs, item.Logs...)
		if len(logs) >= 100 {
			logs = logs[len(logs)-100:]
		}
		params.DataRecord[crontab.ColumnLogs] = logs
	}

	record, err := crontab.NewItem().CheckMap(params.DataRecord)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	if err := response.NewMakeRpcRetErr(l.svcCtx.CrontabModel.UpdateWithParams(record, params), 2); err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return &db.Response{
		Data:    []byte(strconv.FormatBool(true)),
		License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string),
	}, nil
}
