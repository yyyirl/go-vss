/**
 * @Author:         yi
 * @Description:    upload
 * @Version:        1.0.0
 * @Date:           2025/4/24 9:53
 */

package file

import (
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"

	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type Upload struct {
}

func NewUpload() *Upload {
	return new(Upload)
}

func (u *Upload) File(data multipart.File, dir, filename string, useAbsPath bool) (interface{}, *response.HttpErr) {
	var ext = strings.Trim(path.Ext(filename), ".")
	if ext == "" {
		ext = "known"
	}

	var (
		fileDir  = path.Join(dir, ext)
		filePath = path.Join(fileDir, filename)
	)
	_ = os.Remove(filePath)
	if err := functions.MakeDir(fileDir); err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0002)
	}

	root, err := os.Getwd()
	if err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0002)
	}

	dst, err := os.Create(path.Join(root, filePath))
	if err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0002)
	}

	defer dst.Close()

	if _, err := io.Copy(dst, data); err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0002)
	}

	if useAbsPath {
		filePath, err := filepath.Abs(filePath)
		if err != nil {
			return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0002)
		}

		return filePath, nil
	}

	return filePath, nil
}
