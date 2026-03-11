package configservicelogic

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/video-projects"
)

type VideoProjectUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVideoProjectUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VideoProjectUpdateLogic {
	return &VideoProjectUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VideoProjectUpdateLogic) VideoProjectUpdate(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	list, err := l.svcCtx.VideoProjectsModel.List(params)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	record, err := videoProjects.NewItem().CheckMap(params.DataRecord)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	if err := response.NewMakeRpcRetErr(l.svcCtx.VideoProjectsModel.UpdateWithParams(record, params), 2); err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	var channelIds []uint64
	for _, item := range params.Data {
		if item.Column == videoProjects.ColumnChannelUniqueIds {
			channelUniqueIds, ok := item.Value.([]interface{})
			if ok {
				for _, v := range channelUniqueIds {
					if tmp, ok := v.(float64); ok {
						channelIds = append(channelIds, uint64(tmp))
					}
				}
			}
		}
	}

	if len(channelIds) > 0 {
		var ids []interface{}
		for _, item := range list {
			ids = append(ids, item.ID)
		}

		list, err := l.svcCtx.VideoProjectsModel.List(&orm.ReqParams{
			Conditions: []*orm.ConditionItem{{Column: videoProjects.ColumnId, Values: ids, Operator: "notin"}},
			All:        true,
		})
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		for _, item := range list {
			v, err := item.ConvToItem()
			if err != nil {
				return nil, response.NewMakeRpcRetErr(err, 2)
			}

			var (
				exists           = false
				channelUniqueIds []uint64
			)
			for _, val := range v.ChannelUniqueIds {
				if functions.Contains(val, channelIds) {
					exists = true
					continue
				}
				channelUniqueIds = append(channelUniqueIds, val)
			}

			if exists {
				var v = "[]"
				if len(channelUniqueIds) > 0 {
					v, err = functions.ToString(channelUniqueIds)
					if err != nil {
						return nil, response.NewMakeRpcRetErr(err, 2)
					}
				}

				if err := l.svcCtx.VideoProjectsModel.UpdateWithColumns(
					videoProjects.ColumnId,
					item.ID,
					map[string]interface{}{videoProjects.ColumnChannelUniqueIds: v},
					params,
				); err != nil {
					return nil, response.NewMakeRpcRetErr(err, 2)
				}
			}
		}
	}

	return &db.Response{
		Data:    []byte(strconv.FormatBool(true)),
		License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string),
	}, nil
}
