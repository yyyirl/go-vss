package items

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/common/opt"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/system-operation-logs"
)

type DeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLogic {
	return &DeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteLogic) Delete(req *orm.ReqParams) *response.HttpErr {
	// 日志记录
	opt.NewSystemOperationLogs(l.svcCtx.RpcClients).Make(l.ctx, systemOperationLogs.Types[systemOperationLogs.TypeDepartmentDelete], req)

	_, err := response.NewRpcToHttpResp[*deviceservice.Response, bool]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(req)
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.DeviceDelete(l.ctx, data)
		},
	)

	if err != nil {
		return err
	}

	// if len(res.Data) > 0 {
	// 	// 停止流
	// 	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode}).HttpPostJsonResJson(
	// 		fmt.Sprintf("http://%s/api/video/stream/stop", l.svcCtx.Config.VssHttpTarget),
	// 		map[string]interface{}{"streamNames": res.Data},
	// 		nil,
	// 	); err != nil {
	// 		functions.LogError("删除设备 停止流失败, err: ", err)
	// 	}
	// }

	return nil
}
