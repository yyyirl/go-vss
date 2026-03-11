// @Title        opt
// @Description  system_operation_logs
// @Create       yirl 2025/4/7 17:45

package opt

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/common/client"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/system-operation-logs"
)

type SystemOperationLogs struct {
	rpcClients *client.GRPCClients
}

func NewSystemOperationLogs(rpcClients *client.GRPCClients) *SystemOperationLogs {
	return &SystemOperationLogs{
		rpcClients: rpcClients,
	}
}

func (s *SystemOperationLogs) Make(ctx context.Context, Type systemOperationLogs.Type, data interface{}) {
	go func(ctx context.Context) {
		b, err := functions.JSONMarshal(data)
		if err != nil {
			functions.LogError("日志数据写入 data序列化失败")
			return
		}

		if _, err := response.NewRpcToHttpResp[*backendservice.Response, uint64]().Parse(
			func() (*backendservice.Response, error) {
				data, err := structpb.NewStruct(
					(&systemOperationLogs.Item{
						SystemOperationLogs: &systemOperationLogs.SystemOperationLogs{
							Type:   systemOperationLogs.Types[Type],
							Userid: uint64(contextx.GetCtxUserid(ctx)),
							Data:   string(b),
							IP:     contextx.GetCtxIP(ctx),
							Mac:    contextx.GetCtxMAC(ctx),
						},
					}).ToMap(),
				)
				if err != nil {
					return nil, err
				}

				_ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				return s.rpcClients.Backend.SystemOperationLogCreate(_ctx, &backendservice.MapReq{
					Data: data,
				})
			},
		); err != nil {
			functions.LogError("日志创建失败, err:", err)
		}
	}(ctx)
}
