package login

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/common/opt"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/admins"
	systemOperationLogs "skeyevss/core/repositories/models/system-operation-logs"
	"skeyevss/core/tps"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (interface{}, *response.HttpErr) {
	var (
		username = functions.Trim(req.Username)
		password = functions.Trim(req.Password)
	)

	if username == "" {
		return nil, response.MakeError(response.NewHttpRespMessage().Str("用户名不能为空"), localization.M4)
	}

	if password == "" {
		return nil, response.MakeError(response.NewHttpRespMessage().Str("密码不能为空"), localization.M5)
	}

	// 日志记录
	opt.NewSystemOperationLogs(l.svcCtx.RpcClients).Make(l.ctx, systemOperationLogs.Types[systemOperationLogs.TypeLogin], req)

	res, err := response.NewRpcToHttpResp[*backendservice.Response, *admins.Item]().Parse(
		func() (*backendservice.Response, error) {
			return l.svcCtx.RpcClients.Backend.Login(
				l.ctx,
				&backendservice.LoginReq{
					Username: req.Username,
					Password: req.Password,
				},
			)
		},
	)
	if err != nil {
		return nil, err
	}

	if !functions.ValidatePwd(res.Data.Password, password) {
		return nil, response.MakeError(response.NewHttpRespMessage().Str("密码校验失败"), localization.M1)
	}

	if res.Data.Super != 1 && res.Data.State != 1 {
		return nil, response.MakeError(response.NewHttpRespMessage().Str("账号未启用"), localization.M2)
	}

	if res.Data.IsDel != 0 {
		return nil, response.MakeError(response.NewHttpRespMessage().Str("账号已被删除"), localization.M2)
	}

	return l.makeToken(l.svcCtx, res.Data, req.Remember)
}

func (l *LoginLogic) makeToken(svcCtx *svc.ServiceContext, row *admins.Item, remember bool) (interface{}, *response.HttpErr) {
	// 生成token
	var (
		now    = functions.NewTimer().NowMilli()
		expire = time.Duration(now + svcCtx.Config.Auth.LoginExpire)
	)
	if remember {
		expire = time.Duration(now + svcCtx.Config.Auth.LoginRememberExpire)
	}

	token, err := functions.MakeTokenVASE(
		svcCtx.Config.Auth.AesKey,
		expire,
		tps.TokenItem{
			Userinfo: map[string]interface{}{
				admins.ColumnId:       row.ID,
				admins.ColumnNickname: row.Nickname,
				admins.ColumnEmail:    row.Email,
				admins.ColumnSuper:    row.Super,
				admins.ColumnRemark:   row.Remark,
				admins.ColumnAvatar:   row.Avatar,
				admins.ColumnDepIds:   row.DepIds,
				"now":                 now,
			},
		},
	)
	if err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0002)
	}

	return map[string]interface{}{
		"expire":              expire,
		"token":               token,
		admins.ColumnId:       row.ID,
		admins.ColumnNickname: row.Nickname,
		admins.ColumnUsername: row.Username,
		admins.ColumnEmail:    row.Email,
		admins.ColumnAvatar:   row.Avatar,
		admins.ColumnDepIds:   row.DepIds,
	}, nil
}
