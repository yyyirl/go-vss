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

type InitializeSetPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitializeSetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitializeSetPasswordLogic {
	return &InitializeSetPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 初始化密码

func (l *InitializeSetPasswordLogic) InitializeSetPassword(in *db.AdminPasswordReq) (*db.Response, error) {
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

	if data.Username != l.svcCtx.Config.Accounts.BackendUsername && data.Username != l.svcCtx.Config.Accounts.BackendSuperUsername {
		return nil, response.NewMakeRpcRetErr(errors.New("非法操作"), 2)
	}

	var password = l.svcCtx.Config.Accounts.BackendPassword
	if data.Username == l.svcCtx.Config.Accounts.BackendSuperUsername {
		password = l.svcCtx.Config.Accounts.BackendSuperPassword
	}

	if !functions.ValidatePwd(data.Password, password) {
		return nil, response.NewMakeRpcRetErr(errors.New("非法操作"), 2)
	}

	// 更新密码
	newPassword, _ := functions.GeneratePwd(in.Password)
	return nil, response.NewMakeRpcRetErr(
		l.svcCtx.AdminsModel.UpdateWithParams(
			map[string]interface{}{
				admins.ColumnPassword: newPassword,
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
