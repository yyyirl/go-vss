package tool

import (
	"context"
	"path"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type UploadBase64FileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadBase64FileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadBase64FileLogic {
	return &UploadBase64FileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadBase64FileLogic) UploadBase64File(req *types.Base64FileUploadReq) (string, *response.HttpErr) {
	// 保存到本地
	_, fullPath, err := functions.MakeBase64Image(path.Join(l.svcCtx.Config.SaveFileDir, "upload", "base64"), req.Stream)
	if err != nil {
		return "", response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010)
	}

	return fullPath, nil
}
