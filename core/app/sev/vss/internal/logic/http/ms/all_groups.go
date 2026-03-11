package ms

import (
	"context"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/vss/internal/pkg/ms"
	"skeyevss/core/app/sev/vss/internal/types"
	ctypes "skeyevss/core/common/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

var (
	_ types.HttpRHandleLogic[*AllGroupsLogic, types.DCReq] = (*AllGroupsLogic)(nil)

	VAllGroupsLogic = new(AllGroupsLogic)
)

type AllGroupsLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *AllGroupsLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *AllGroupsLogic {
	return &AllGroupsLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *AllGroupsLogic) Path() string {
	return "/ms/all_groups"
}

func (l *AllGroupsLogic) DO(req types.DCReq) *types.HttpResponse {
	var msIds []uint64
	if req.DeviceUniqueId != "" && req.ChannelUniqueId != "" {
		res, err := response.NewRpcToHttpResp[*deviceservice.Response, *ctypes.DeviceChannel]().Parse(
			func() (*deviceservice.Response, error) {
				return l.svcCtx.RpcClients.Device.DeviceChannel(l.ctx, &deviceservice.DeviceChannelReq{
					ChannelUniqueId: req.ChannelUniqueId,
					DeviceUniqueId:  req.DeviceUniqueId,
				})
			},
		)
		if err != nil {
			return &types.HttpResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Str(err.Error), localization.M0010),
			}
		}

		if res.Data.Device == nil || res.Data.Channel == nil {
			return &types.HttpResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Str("设备获取失败"), localization.M0010),
			}
		}

		msIds = res.Data.Device.MSIds
	} else if req.MsID > 0 {
		msIds = []uint64{req.MsID}
	}

	var msNode = ms.New(l.ctx, l.svcCtx).VoteNode(msIds)
	if msNode == nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("未设置流媒体源"), localization.M0010),
		}
	}

	list, err1 := ms.New(l.ctx, l.svcCtx).GetStreamAllGroup(msNode.Address)
	if err1 != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(err1), localization.M0010),
		}
	}

	return &types.HttpResponse{Data: list}
}
