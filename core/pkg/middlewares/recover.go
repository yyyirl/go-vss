/**
 * @Author:         yi
 * @Description:    recover
 * @Version:        1.0.0
 * @Date:           2022/10/11 15:48
 */
package middlewares

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"skeyevss/core/constants"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

func (m *MW) recover(w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		var (
			broken = string(debug.Stack())
			info   = fmt.Sprintf(
				"error:%+v\nbroken: %s\nauthorization: %s\nurl: %s\nrequest method: %s, \nstack: %s",
				err,
				broken,
				r.Header.Get(constants.HEADER_AUTHORIZATION),
				r.RequestURI,
				r.Method,
				string(debug.Stack()),
			)
		)

		functions.LogError("recover info:", info)
		if m.MailFunc != nil {
			go m.MailFunc()(info, broken)
		}

		response.New().RequestError(r.Context(), w, response.MakeError(response.NewHttpRespMessage().Str(info), localization.M0007))

	}
}
