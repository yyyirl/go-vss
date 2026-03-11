package channels

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/common"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/db/client/configservice"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/common/stream"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
	videoProjects "skeyevss/core/repositories/models/video-projects"
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

func (l *ListLogic) List(req *orm.ReqParams) (interface{}, *response.HttpErr) {
	var departmentIds = common.DepartmentIds(l.ctx)
	if departmentIds != nil {
		req.Conditions = append(req.Conditions, &orm.ConditionItem{
			Column:   channels.ColumnDepIds,
			Original: orm.NewExternalDB(l.svcCtx.Config.SevBase.DatabaseType).MakeCaseJSONContainsCondition(channels.ColumnDepIds, functions.SliceToSliceAny(departmentIds)),
		})
	}

	res, err := response.NewRpcToHttpResp[*deviceservice.Response, *response.ListWithMapResp[[]*channels.Item, string]]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(req)
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.ChannelList(l.ctx, data)
		},
	)
	if err != nil {
		return nil, err
	}

	var ids []uint64
	for _, item := range res.Data.List {
		ids = append(ids, item.ID)
	}

	if videoProjectRes, _ := response.NewRpcToHttpResp[*configservice.Response, *response.ListResp[[]*videoProjects.Item]]().Parse(
		func() (*configservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{All: true})
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Config.VideoProjectList(l.ctx, data)
		},
	); len(videoProjectRes.Data.List) > 0 {
		var maps = make(map[uint64][]string)
		for _, item := range videoProjectRes.Data.List {
			for _, id := range item.ChannelUniqueIds {
				if functions.Contains(id, ids) {
					maps[id] = strings.Split(item.Plans, "")
				}
			}
		}

		var data = make(map[uint64]map[int][][2]string)
		for id, item := range maps {
			if len(item) != 168 {
				continue
			}

			data[id] = stream.NewVideoPlain().Views([168]string(item))
		}

		res.Data.Ext = map[string]interface{}{
			"plans": data,
		}
	}

	return res.Data, nil
}
