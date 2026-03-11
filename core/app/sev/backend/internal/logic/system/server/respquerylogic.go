package server

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/constants"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type RespQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRespQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RespQueryLogic {
	return &RespQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RespQueryLogic) RespQuery(req *types.RespQueryReq) (interface{}, *response.HttpErr) {
	if req.Code == "" {
		return "", response.MakeError(response.NewHttpRespMessage().Str("code不能为空"), localization.M00280)
	}

	val, err := functions.NewCrypto([]byte(constants.RESPONSE_AES_KEY)).Decrypt(req.Code)
	if err != nil {
		return "", response.MakeError(response.NewHttpRespMessage().Str("code不能为空"), localization.M0026)
	}

	return strings.Split(strings.Replace(val, "\\/", "/", -1), "\n"), nil
}
