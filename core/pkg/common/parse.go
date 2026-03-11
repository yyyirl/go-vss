// @Title        parse
// @Description  main
// @Create       yiyiyi 2025/7/11 15:18

package common

import (
	"io"
	"net/http"
	"skeyevss/core/pkg/functions"
)

func Parse(r *http.Request, data interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer func() {
		_ = r.Body.Close()
	}()

	return functions.JSONUnmarshal(body, &data)
}
