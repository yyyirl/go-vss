package functions

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	resty "github.com/go-resty/resty/v2"
)

var RestyDebug bool

type ResponseJson struct {
	Code uint32      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type (
	RestyConfig struct {
		Mode,
		Referer string
		RetryCount int
	}

	Resty struct {
		ctx    context.Context
		config *RestyConfig

		client *resty.Request
	}

	restyLogger struct {
		mode string
	}
)

func (l *restyLogger) Errorf(format string, v ...interface{}) {
	LogError(fmt.Sprintf(format+Caller(6), v...))
}

func (l *restyLogger) Warnf(format string, v ...interface{}) {
	LogAlert(fmt.Sprintf(format+Caller(6), v...))
}

func (l *restyLogger) Debugf(format string, v ...interface{}) {
	LogInfo(fmt.Sprintf(format+Caller(6), v...))
}

func NewResty(ctx context.Context, config *RestyConfig) *Resty {
	var client = resty.New()
	// client.SetTimeout(time.Second * httpClientTimeOut)
	if config.RetryCount > 0 {
		client.SetRetryCount(config.RetryCount)
	}
	client.SetRetryCount(1)
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.SetLogger(&restyLogger{mode: config.Mode})

	return &Resty{
		ctx:    ctx,
		config: config,
		client: client.R().SetContext(ctx).SetDebug(config.Mode == "dev" && RestyDebug).SetHeader("Referer", config.Referer),
	}
}

// --------------------------------------------------- get json

func (s *Resty) HttpGetResJson(url string, queryParams map[string]string, result interface{}) (res *resty.Response, err error) {
	res, err = s.client.
		SetQueryParams(queryParams).
		SetHeader("Accept", "application/json").
		// SetResult(result).
		Get(url)

	if err == nil {
		if err := JSONUnmarshal(res.Body(), result); err != nil {
			return res, err
		}
	}

	return res, err
}

func (s *Resty) HttpGet(url string, queryParams map[string]string) (res *resty.Response, err error) {
	return s.client.SetQueryParams(queryParams).Get(url)
}

// --------------------------------------------------- post res json

func (s *Resty) HttpPostFormResJson(url string, formData map[string]string, result interface{}) (res *resty.Response, err error) {
	return s.HttpSendFormResJson(url, "POST", formData, result)
}

func (s *Resty) HttpPutFormResJson(url string, formData map[string]string, result interface{}) (res *resty.Response, err error) {
	return s.HttpSendFormResJson(url, "PUT", formData, result)
}

func (s *Resty) HttpPatchFormResJson(url string, formData map[string]string, result interface{}) (res *resty.Response, err error) {
	return s.HttpSendFormResJson(url, "PATCH", formData, result)
}

func (s *Resty) HttpDeleteFormResJson(url string, formData map[string]string, result interface{}) (res *resty.Response, err error) {
	return s.HttpSendFormResJson(url, "DELETE", formData, result)
}

func (s *Resty) HttpOptionsFormResJson(url string, formData map[string]string, result interface{}) (res *resty.Response, err error) {
	return s.HttpSendFormResJson(url, "OPTIONS", formData, result)
}

func (s *Resty) HttpHeadFormResJson(url string, formData map[string]string, result interface{}) (res *resty.Response, err error) {
	return s.HttpSendFormResJson(url, "HEAD", formData, result)
}

func (s *Resty) HttpPostJsonResJson(url string, formData map[string]interface{}, result interface{}) (res *resty.Response, err error) {
	return s.HttpSendJsonResJson(url, "POST", formData, result)
}

func (s *Resty) HttpPutJsonResJson(url string, formData map[string]string, result interface{}) (res *resty.Response, err error) {
	return s.HttpSendJsonResJson(url, "PUT", formData, result)
}

func (s *Resty) HttpPatchJsonResJson(url string, formData map[string]string, result interface{}) (res *resty.Response, err error) {
	return s.HttpSendJsonResJson(url, "PATCH", formData, result)
}

func (s *Resty) HttpDeleteJsonResJson(url string, formData map[string]string, result interface{}) (res *resty.Response, err error) {
	return s.HttpSendJsonResJson(url, "DELETE", formData, result)
}

func (s *Resty) HttpOptionsJsonResJson(url string, formData map[string]string, result interface{}) (res *resty.Response, err error) {
	return s.HttpSendJsonResJson(url, "OPTIONS", formData, result)
}

func (s *Resty) HttpHeadJsonResJson(url string, formData map[string]string, result interface{}) (res *resty.Response, err error) {
	return s.HttpSendJsonResJson(url, "HEAD", formData, result)
}

func (s *Resty) HttpSendFormResJson(url, method string, formData map[string]string, result interface{}) (res *resty.Response, err error) {
	if url == "" {
		return nil, errors.New("url is empty")
	}

	req := s.client.
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").
		SetFormData(formData)
	// .SetResult(result)

	switch strings.ToLower(method) {
	case "post":
		res, err = req.Post(url)
	case "put":
		res, err = req.Put(url)
	case "patch":
		res, err = req.Patch(url)
	case "delete":
		res, err = req.Delete(url)
	case "options":
		res, err = req.Options(url)
	default:
		res, err = req.Head(url)
	}

	if err == nil {
		if err := JSONUnmarshal(res.Body(), result); err != nil {
			return res, err
		}
	}

	return res, err
}

// HttpSendJsonResJson send json and response json
func (s *Resty) HttpSendJsonResJson(url, method string, body interface{}, result interface{}) (res *resty.Response, err error) {
	if url == "" {
		return nil, errors.New("url is empty")
	}

	var req = s.client.SetHeader(
		"Content-Type",
		"application/json",
	).SetHeader(
		"Accept",
		"application/json",
	).SetBody(body)
	// .SetResult(result)

	if s.config.Mode == "dev" {
		b, _ := JSONMarshal(body)
		LogInfo("请求地址: ", url, "; 请求参数: ", string(b))
	}

	switch strings.ToLower(method) {
	case "post":
		res, err = req.Post(url)
	case "put":
		res, err = req.Put(url)
	case "patch":
		res, err = req.Patch(url)
	case "delete":
		res, err = req.Delete(url)
	case "options":
		res, err = req.Options(url)
	default:
		res, err = req.Head(url)
	}

	if err == nil {
		if err := JSONUnmarshal(res.Body(), result); err != nil {
			return res, err
		}
	}

	return res, err
}
