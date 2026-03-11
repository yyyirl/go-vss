package setting

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
)

type ConfLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfLogic {
	return &ConfLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfLogic) Conf() interface{} {
	var data = map[string]interface{}{
		"title":   l.svcCtx.Settings().Content.WebManageTitle,
		"website": l.svcCtx.Settings().Content.Website,
	}
	if l.svcCtx.Config.UseShowcaseAccount {
		data["showcase-username"] = l.svcCtx.Config.Accounts.BackendShowcaseUsername
		data["showcase-password"] = l.svcCtx.Config.Accounts.BackendShowcasePassword
	}

	return data
}
