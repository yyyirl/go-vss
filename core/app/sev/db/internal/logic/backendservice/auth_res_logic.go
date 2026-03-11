package backendservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/common/types"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/admins"
	"skeyevss/core/repositories/models/departments"
	"skeyevss/core/repositories/models/roles"
)

type AuthResLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuthResLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthResLogic {
	return &AuthResLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AuthResLogic) AuthRes(_ *db.EmptyRequest) (*db.Response, error) {
	// 管理员
	adminList, queryErr := l.svcCtx.AdminsModel.List(&orm.ReqParams{All: true})
	if queryErr != nil {
		return nil, response.NewMakeRpcRetErr(queryErr, 2)
	}

	var adminRecords []*admins.Item
	for _, item := range adminList {
		v, err := item.ConvToItem()
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		adminRecords = append(adminRecords, v)
	}

	// 部门
	departmentsList, queryErr := l.svcCtx.DepartmentsModel.List(&orm.ReqParams{All: true})
	if queryErr != nil {
		return nil, response.NewMakeRpcRetErr(queryErr, 2)
	}

	var departmentRecords []*departments.Item
	for _, item := range departmentsList {
		v, err := item.ConvToItem()
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		departmentRecords = append(departmentRecords, v)
	}

	// 角色
	roleList, queryErr := l.svcCtx.RolesModel.List(&orm.ReqParams{All: true})
	if queryErr != nil {
		return nil, response.NewMakeRpcRetErr(queryErr, 2)
	}

	var roleRecords []*roles.Item
	for _, item := range roleList {
		v, err := item.ConvToItem()
		if err != nil {
			return nil, response.NewMakeRpcRetErr(err, 2)
		}

		roleRecords = append(roleRecords, v)
	}

	return response.NewRpcResp[*db.Response]().Make(&types.AuthRes{
		Admins:      adminRecords,
		Departments: departmentRecords,
		Roles:       roleRecords,
	}, 3, func(data []byte) *db.Response {
		return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	})
}
