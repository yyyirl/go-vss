package backendservicelogic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/admins"
	"skeyevss/core/repositories/models/system-operation-logs"
)

type SystemOperationLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSystemOperationLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SystemOperationLogsLogic {
	return &SystemOperationLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SystemOperationLogsLogic) SystemOperationLogs(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	var (
		conditions []*orm.ConditionItem
		usernames  []string
	)
	for _, item := range params.Conditions {
		if item.Column == "username" {
			if v, ok := item.Value.(string); ok {
				usernames = append(usernames, v)
				continue
			}

			return nil, response.NewMakeRpcRetErr(errors.New("字段类型错误"), 2)
		}
		conditions = append(conditions, item)
	}

	// 获取管理员id
	if len(usernames) > 0 {
		adminList, err := l.svcCtx.AdminsModel.List(&orm.ReqParams{
			Limit: len(usernames),
			Conditions: []*orm.ConditionItem{
				{
					Column: admins.ColumnUsername,
					Values: functions.SliceToSliceAny(usernames),
				},
			},
		})
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		if len(adminList) <= 0 {
			return response.NewRpcResp[*db.Response]().Make(response.NewListResp[[]*systemOperationLogs.Item]().Empty(), 3, func(data []byte) *db.Response {
				return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
			})
		}

		var userIds []uint64
		for _, item := range adminList {
			userIds = append(userIds, item.ID)
		}
		userIds = functions.ArrUnique(userIds)

		conditions = append(conditions, &orm.ConditionItem{
			Column: systemOperationLogs.ColumnUserid,
			Values: functions.SliceToSliceAny(userIds),
		})
	}

	params.Conditions = conditions
	// 获取总数
	count, queryErr := l.svcCtx.SystemOperationLogsModel.Count(params)
	if queryErr != nil {
		return nil, response.NewMakeRpcRetErr(queryErr, 2)
	}

	if count <= 0 {
		return response.NewRpcResp[*db.Response]().Make(response.NewListResp[[]*systemOperationLogs.Item]().Empty(), 3, func(data []byte) *db.Response {
			return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
		})
	}

	if len(params.Orders) <= 0 {
		params.Orders = []*orm.OrderItem{
			{Column: systemOperationLogs.ColumnCreatedAt, Value: orm.OrderDesc},
		}
	}
	// 获取列表
	list, queryErr := l.svcCtx.SystemOperationLogsModel.List(params)
	if queryErr != nil {
		return nil, response.NewMakeRpcRetErr(queryErr, 2)
	}

	var (
		records []*systemOperationLogs.Item
		userIds []uint64
	)
	for _, item := range list {
		v, err := item.ConvToItem()
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		if item.Userid != 0 {
			userIds = append(userIds, item.Userid)
		}
		records = append(records, v)
	}

	// 管理员信息
	if len(userIds) > 0 {
		userIds = functions.ArrUnique(userIds)
		adminList, err := l.svcCtx.AdminsModel.List(&orm.ReqParams{
			Limit: len(userIds),
			Conditions: []*orm.ConditionItem{
				{
					Column: admins.ColumnId,
					Values: functions.SliceToSliceAny(userIds),
				},
			},
		})
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		var maps = make(map[uint64]string)
		for _, item := range adminList {
			maps[item.ID] = item.Username
		}

		for _, item := range records {
			if v, ok := maps[item.Userid]; ok {
				item.Username = v
			}
		}
	}

	return response.NewRpcResp[*db.Response]().Make(&response.ListResp[[]*systemOperationLogs.Item]{
		List:  records,
		Count: count,
	}, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
