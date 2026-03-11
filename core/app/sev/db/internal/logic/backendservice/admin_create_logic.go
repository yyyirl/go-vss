package backendservicelogic

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/admins"
)

type AdminCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateLogic {
	return &AdminCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminCreateLogic) AdminCreate(in *db.MapReq) (*db.Response, error) {
	// 管理员创建
	record, err := admins.NewItem().MapToModel(in.Data.AsMap())
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	data, err := record.ConvToModel(func(item *admins.Item) *admins.Item {
		item.Password, _ = functions.GeneratePwd(item.Password)
		return item
	})
	if err != nil || data == nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	res, err := l.svcCtx.AdminsModel.Add(*data)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	return &db.Response{
		Data:    []byte(strconv.Itoa(int(res.ID))),
		License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string),
	}, nil
}
