package tool

import (
	"net/http"

	"skeyevss/core/app/sev/backend/internal/logic/tool"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

func UploadFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()
		_ = r.ParseMultipartForm(svcCtx.Config.MaxBytes << 20)
		file, handler, err := r.FormFile("file")
		if err != nil {
			response.New().RequestError(ctx, w, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001))
			return
		}

		var (
			fileName       = handler.Filename
			headerFilename = r.Header.Get("X-Filename")
			useAbsPath     = r.PostFormValue("abs") == "1"
		)
		if headerFilename != "" {
			fileName = headerFilename
		}

		defer func() {
			_ = file.Close()
		}()
		resp, err1 := tool.NewUploadFileLogic(ctx, svcCtx).UploadFile(file, fileName, useAbsPath)
		if err1 != nil {
			response.New().RequestError(ctx, w, err1)
			return
		}

		response.New().Success(ctx, w, resp)
	}
}
