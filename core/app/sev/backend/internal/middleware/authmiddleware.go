package middleware

import (
	"context"
	"net/http"

	"skeyevss/core/app/sev/backend/internal/config"
	"skeyevss/core/common/types"
	"skeyevss/core/constants"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/middlewares"
	"skeyevss/core/repositories/models/admins"
)

type AuthMiddleware struct {
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

func (m *AuthMiddleware) permissions(userid uint, authRes *types.AuthRes) (*admins.Item, []string, *localization.Item) {
	if userid <= 0 {
		return nil, nil, localization.M0008
	}

	// 当前登录账户权限
	var adminRow *admins.Item
	for _, item := range authRes.Admins {
		if item.ID == uint64(userid) {
			adminRow = item
		}
	}
	if adminRow == nil {
		return nil, nil, localization.M0008
	}

	if adminRow.Super >= 1 {
		return adminRow, nil, nil
	}

	if len(adminRow.DepIds) <= 0 {
		return nil, nil, localization.M0011
	}

	var roleIds []uint64
	for _, item := range authRes.Departments {
		if item.State == 0 {
			continue
		}

		if !functions.Contains(item.ID, adminRow.DepIds) {
			continue
		}
		roleIds = append(roleIds, item.RoleIds...)
	}

	var permissionIds []string
	for _, item := range authRes.Roles {
		if functions.Contains(item.ID, roleIds) && item.State == 1 {
			permissionIds = append(permissionIds, item.PermissionUniqueIds...)
		}
	}

	return adminRow, permissionIds, nil
}

func (m *AuthMiddleware) Handle(c config.Config, next http.HandlerFunc, authRes *types.AuthRes, _ string) http.HandlerFunc {
	return middlewares.New(
		middlewares.Conf{
			AesKey:            c.Auth.AesKey,
			Secret:            c.Auth.JwtSecret,
			Expire:            c.Auth.LoginExpire,
			TokenVerification: true,
			MailCall: func() func(info, broken string) {
				return recoverCallback(c)
			},
			CustomerCall: func(ctx context.Context, r *http.Request) (context.Context, *localization.Item) {
				var userid = contextx.GetCtxUserid(ctx)
				if userid <= 0 {
					return ctx, localization.M0008
				}

				// 当前登录账户权限
				adminRow, permissionIds, err := m.permissions(userid, authRes)
				if err != nil {
					return ctx, err
				}

				var showcase = false
				// 检测是否更新密码
				if c.InitializeSetPassword {
					for _, item := range authRes.Admins {
						if item.ID == uint64(userid) {
							if item.Username == c.Accounts.BackendShowcaseUsername && c.UseShowcaseAccount {
								showcase = true
							}

							if item.Username == c.Accounts.BackendUsername {
								if functions.ValidatePwd(item.Password, c.Accounts.BackendPassword) {
									// 提示更新密码
									ctx = context.WithValue(ctx, constants.CTX_NEEDED_RESET_PWD, true)
								}
							} else if item.Username == c.Accounts.BackendSuperUsername {
								if functions.ValidatePwd(item.Password, c.Accounts.BackendSuperPassword) {
									// 提示更新密码
									ctx = context.WithValue(ctx, constants.CTX_NEEDED_RESET_PWD, true)
								}
							}
							break
						}
					}
				}

				ctx = context.WithValue(ctx, constants.CTX_DEP_IDS, adminRow.DepIds)
				ctx = context.WithValue(ctx, constants.CTX_SUPER_STATE, adminRow.Super)
				ctx = context.WithValue(ctx, constants.CTX_PERMISSION_IDS, permissionIds)
				ctx = context.WithValue(ctx, constants.CTX_SHOWCASE, showcase)

				return ctx, nil
			},
		},
		next,
	)
}
