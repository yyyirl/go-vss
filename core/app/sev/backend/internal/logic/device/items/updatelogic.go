package items

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/common/opt"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/devices"
	systemOperationLogs "skeyevss/core/repositories/models/system-operation-logs"
)

type UpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateLogic) Update(req *orm.ReqParams) *response.HttpErr {
	// 日志记录
	opt.NewSystemOperationLogs(l.svcCtx.RpcClients).Make(l.ctx, systemOperationLogs.Types[systemOperationLogs.TypeDepartmentUpdate], req)

	rowRes, err := response.NewRpcToHttpResp[*deviceservice.Response, *devices.Item]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(req)
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.DeviceRow(l.ctx, data)
		},
	)
	if err != nil {
		return err
	}

	if _, err := response.NewRpcToHttpResp[*deviceservice.Response, bool]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(req)
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.DeviceUpdate(l.ctx, data)
		},
	); err != nil {
		return err
	}

	var rq = l.svcCtx.RemoteReq(l.ctx)
	for _, item := range req.Data {
		if item.Column == devices.ColumnSubscription {
			if v, ok := item.Value.(string); ok {
				var subscription = (devices.Devices{}).ConvSubscription(v)
				if !(subscription.Catalog || subscription.EmergencyCall || subscription.Location || subscription.PTZ) {
					break
				}

				// 发送订阅消息
				if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode, Referer: rq.Referer}).HttpPostJsonResJson(
					fmt.Sprintf("%s/api/gbs/subscription", rq.VssHttpUrlInternal),
					map[string]interface{}{
						"subscription":   subscription,
						"deviceUniqueId": rowRes.Data.DeviceUniqueId,
					},
					nil,
				); err != nil {
					return response.MakeError(response.NewHttpRespMessage().Err(err), localization.MR1003)
				}
			}

			break
		}
	}

	return nil
}
