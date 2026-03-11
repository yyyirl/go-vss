package devices

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	cTypes "skeyevss/core/common/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
)

type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListLogic) List(req *types.QueryVideoRecordsReq) (interface{}, *response.HttpErr) {
	var reqParams map[string]interface{}
	if err := functions.ConvInterface(req, &reqParams); err != nil {
		return 0, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0026)
	}

	// 设备信息
	channelMaps, err := response.NewRpcToHttpResp[*deviceservice.Response, map[uint64]*cTypes.ChannelMSRelItem]().Parse(
		func() (*backendservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: channels.ColumnUniqueId, Value: req.ChannelUniqueId},
				},
				All: true,
			})
			if err != nil {
				return nil, response.NewMakeRpcRetErr(err, 2)
			}

			return l.svcCtx.RpcClients.Device.MediaServersWithChannelIds(l.ctx, data)
		},
	)
	if err != nil {
		return "", err
	}

	var maps = make(map[string]map[string]interface{})
	for _, item := range channelMaps.Data {
		if len(item.MSIds) > 0 {
			maps[item.ChannelUniqueId] = map[string]interface{}{
				"address": l.svcCtx.MSVoteNode(item.MSIds).Address,
				"msIds":   item.MSIds,
			}
		}
	}

	var (
		resp response.HttpResp[response.ListWithExtResp[[]interface{}, [][2]string, string]]
		rq   = l.svcCtx.RemoteReq(l.ctx)
	)
	if _, err := functions.NewResty(l.ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode, Referer: rq.Referer}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/device-videos", rq.VssHttpUrlInternal),
		reqParams,
		&resp,
	); err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(fmt.Sprintf("onvif discover设备获取失败, err: %s", err)), localization.M0010)
	}

	if resp.Error != "" {
		return nil, response.MakeError(response.NewHttpRespMessage().Str(resp.Error), localization.MR1008)
	}

	resp.Data.Ext = map[string]interface{}{
		"maps": maps,
	}

	return resp.Data, nil
}
