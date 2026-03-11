package configservicelogic

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/video-projects"
)

type VideoProjectCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVideoProjectCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VideoProjectCreateLogic {
	return &VideoProjectCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VideoProjectCreateLogic) VideoProjectCreate(in *db.MapReq) (*db.Response, error) {
	record, err := videoProjects.NewItem().MapToModel(in.Data.AsMap())
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	data, err := record.ConvToModel(func(item *videoProjects.Item) *videoProjects.Item {
		return item
	})
	if err != nil || data == nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	res, err := l.svcCtx.VideoProjectsModel.Add(*data)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return &db.Response{
		Data:    []byte(strconv.Itoa(int(res.ID))),
		License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string),
	}, nil
}
