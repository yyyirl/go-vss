package response

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"

	"skeyevss/core/constants"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/functions"
)

type (
	HttpRepSuccess struct {
		Data          interface{} `json:"data"`
		Token         string      `json:"token"`
		Logout        bool        `json:"logout"`
		Node          string      `json:"node,optional,omitempty"`
		CallConsuming int64       `json:"callConsuming,optional,omitempty"`
		ResetPwd      bool        `json:"reset-pwd,optional,omitempty"`
	}

	HttpRepError struct {
		Error         string `json:"error,omitempty"`
		Message       string `json:"message,omitempty"`
		Token         string `json:"token,omitempty"`
		Logout        bool   `json:"logout,omitempty"`
		Code          int    `json:"code,omitempty"`
		CallConsuming int64  `json:"callConsuming,optional,omitempty"`
		ResetPwd      bool   `json:"reset-pwd,optional,omitempty"`
	}

	HttpResp[T any] struct {
		ResetPwd      bool     `json:"reset-pwd,optional,omitempty"`
		Logout        bool     `json:"logout,optional,omitempty"`
		Message       string   `json:"message,optional,omitempty"`
		Token         string   `json:"token,optional,omitempty"`
		Code          int      `json:"code,optional,omitempty"`
		Error         string   `json:"error,optional,omitempty"`
		Timestamp     int64    `json:"timestamp,optional,omitempty"`
		Node          string   `json:"node,optional,omitempty"`
		Version       string   `json:"version,optional,omitempty"`
		Errors        []string `json:"errors,optional,omitempty"`
		Data          T        `json:"data,optional,omitempty"`
		CallConsuming float64  `json:"cc,optional,omitempty"`
	}
)

func (rs *HttpRepError) ToMap() *HttpResp[any] {
	var (
		now = functions.NewTimer().NowMilli()
		rep = HttpResp[any]{
			Error:         rs.Error,
			Timestamp:     now,
			CallConsuming: (float64(now) - float64(rs.CallConsuming)) / 1000,
		}
	)

	if rs.Token != "" {
		rep.Token = rs.Token
	}

	if rs.Message != "" {
		rep.Message = rs.Message
	}

	if rs.Code > 0 {
		rep.Code = rs.Code
	}

	if version := functions.GetEnvDefault("SKEYEVSS_VERSION", ""); version != "" {
		rep.Version = version
	}

	if rs.Error != "" {
		if constants.ENV != constants.ENV_PRODUCTION {
			rep.Errors = strings.Split(rs.Error, "\n")
		}
	}

	rep.Node = functions.ServerNode()
	return &rep
}

func (rs *HttpRepSuccess) ToMap() *HttpResp[any] {
	var (
		now  = functions.NewTimer().NowMilli()
		data = &HttpResp[any]{
			Timestamp:     functions.NewTimer().NowMilli(),
			CallConsuming: (float64(now) - float64(rs.CallConsuming)) / 1000,
		}
	)

	if rs.Token != "" {
		data.Token = rs.Token
	}

	if rs.Logout {
		data.Logout = rs.Logout
	}

	if rs.ResetPwd {
		data.ResetPwd = rs.ResetPwd
	}

	if version := functions.GetEnvDefault("SKEYEVSS_VERSION", ""); version != "" {
		data.Version = version
	}

	if rs.Data != nil {
		data.Data = rs.Data
	} else {
		data.Data = true
	}

	data.Node = functions.ServerNode()
	return data
}

type (
	HttpRespMessage struct {
	}

	HttpRespMessageContent struct {
		Error error
	}
)

func NewHttpRespMessage() *HttpRespMessage {
	return new(HttpRespMessage)
}

func (r *HttpRespMessage) Err(err error) *HttpRespMessageContent {
	if err != nil {
		return &HttpRespMessageContent{Error: err}
	}

	return &HttpRespMessageContent{Error: errors.New("no input")}
}

func (r *HttpRespMessage) Str(str string, a ...any) *HttpRespMessageContent {
	return &HttpRespMessageContent{Error: errors.New(fmt.Sprintf(str, a...))}
}

type (
	HttpResponse struct{}
	HttpErr      struct {
		Error    string             `json:"error"`
		Message  *localization.Item `json:"message"`
		HttpCode int                `json:"httpCode"` // http code
		Code     int                `json:"code"`
	}
)

func (r *HttpErr) WithCode(code int) *HttpErr {
	r.Code = code
	return r
}

func New() *HttpResponse {
	return new(HttpResponse)
}

func (r *HttpResponse) Success(ctx context.Context, w http.ResponseWriter, data interface{}) {
	httpx.OkJson(
		w,
		(&HttpRepSuccess{
			Data:          data,
			Token:         contextx.GetNewToken(ctx),
			Logout:        contextx.GetLogoutState(ctx),
			CallConsuming: contextx.GetCtxReqStartTime(ctx),
			ResetPwd:      contextx.GetResetPwdState(ctx),
		}).ToMap(),
	)
}

func (r *HttpResponse) err(ctx context.Context, w http.ResponseWriter, responseErr *HttpErr) {
	var (
		language = contextx.GetLanguage(ctx)
		errResp  = &HttpRepError{
			Code:          responseErr.Code,
			Error:         responseErr.Error,
			Token:         contextx.GetNewToken(ctx),
			CallConsuming: contextx.GetCtxReqStartTime(ctx),
		}
	)
	if constants.ENV == constants.ENV_PRODUCTION {
		val, err := functions.NewCrypto([]byte(constants.RESPONSE_AES_KEY)).Encrypt([]byte(strings.Replace(responseErr.Error, "/", "\\/", -1)))
		if err != nil {
			errResp.Error = err.Error()
			return
		}
		errResp.Error = val
	}

	switch language {
	case constants.LANG_EN:
		errResp.Message = responseErr.Message.EN
	default:
		errResp.Message = responseErr.Message.ZH
	}

	httpx.WriteJson(w, responseErr.HttpCode, errResp.ToMap())
}

func (r *HttpResponse) RequestError(ctx context.Context, w http.ResponseWriter, error *HttpErr) {
	r.err(ctx, w, error)
}

// 400

func MakeError(content *HttpRespMessageContent, message *localization.Item) *HttpErr {
	if message != nil {
		if strings.Contains(message.ZH, rpcErrPrefix) {
			var data = strings.Split(message.ZH, rpcErrPrefix)
			message.ZH = data[0]
		}

		if strings.Contains(message.EN, rpcErrPrefix) {
			var data = strings.Split(message.EN, rpcErrPrefix)
			message.EN = data[0]
		}
	}

	return &HttpErr{
		Message:  message,
		Error:    "error: " + content.Error.Error() + "\ncaller: " + functions.CallerFile(2),
		HttpCode: http.StatusBadRequest,
	}
}

// 401

func MakeUnauthorizedError(err error, message *localization.Item) *HttpErr {
	return &HttpErr{
		Message:  message,
		Error:    "error: " + err.Error() + "\ncaller: " + functions.CallerFile(2),
		HttpCode: http.StatusUnauthorized,
	}
}

// 403

func MakeForbiddenError(err error, message *localization.Item) *HttpErr {
	return &HttpErr{
		Message:  message,
		Error:    "error: " + err.Error() + "\ncaller: " + functions.CallerFile(2),
		HttpCode: http.StatusForbidden,
		Code:     http.StatusForbidden,
	}
}
