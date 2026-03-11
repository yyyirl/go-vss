package backendservicelogic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/admins"
)

type AdminPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminPasswordLogic {
	return &AdminPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新密码

func (l *AdminPasswordLogic) AdminPassword(in *db.AdminPasswordReq) (*db.Response, error) {
	var adminId = interceptor.GetAdminId(l.ctx)
	if adminId <= 0 {
		return nil, response.NewMakeRpcRetErr(errors.New("未登录"), 2)
	}

	row, err := l.svcCtx.AdminsModel.RowWithParams(
		&orm.ReqParams{
			Conditions: []*orm.ConditionItem{
				{
					Column: admins.ColumnId,
					Value:  adminId,
				},
			},
		},
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewMakeRpcRetErr(errors.New("账号不存在"), 2)
		}

		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	data, err := row.ConvToItem()
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	if !functions.ValidatePwd(data.Password, in.OldPassword) {
		return nil, response.NewMakeRpcRetErr(errors.New("密码错误"), 2)
	}

	// 更新密码
	password, _ := functions.GeneratePwd(in.Password)
	return nil, response.NewMakeRpcRetErr(
		l.svcCtx.AdminsModel.UpdateWithParams(
			map[string]interface{}{
				admins.ColumnPassword: password,
			},
			&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{
						Column: admins.ColumnId,
						Value:  adminId,
					},
				},
			},
		),
		2,
	)
}
