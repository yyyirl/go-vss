package configservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/response"
)

type VideoProjectRowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVideoProjectRowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VideoProjectRowLogic {
	return &VideoProjectRowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VideoProjectRowLogic) VideoProjectRow(in *db.IDReq) (*db.Response, error) {
	row, err := l.svcCtx.VideoProjectsModel.Row(in.ID)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	data, err := row.ConvToItem()
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return response.NewRpcResp[*db.Response]().Make(data, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
