package tool

import (
	"context"
	"mime/multipart"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/common/file"
	"skeyevss/core/pkg/response"
)

type UploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadFileLogic {
	return &UploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadFileLogic) UploadFile(data multipart.File, filename string, useAbsPath bool) (interface{}, *response.HttpErr) {
	return file.NewUpload().File(data, l.svcCtx.Config.SaveFileDir+"/upload", filename, useAbsPath)
}
