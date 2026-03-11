package sip

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/ghettovoice/gosip/sip"

	"skeyevss/core/app/sev/vss/internal/pkg/rule"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
)

const (
	HeaderUserAgentKey       = "user-agent"
	HeaderWWWAuthenticateKey = "WWW-Authenticate"
	HeaderAuthenticateKey    = "Authorization"
	HeaderExpiresKey         = "Expires"
	HeaderToKey              = "To"
	HeaderContentType        = "Content-Type"
	HeaderContentLength      = "Content-Length"
)

type handler[T types.SipReceiveHandleLogic[T]] struct {
	req    sip.Request
	tx     sip.ServerTransaction
	svcCtx *types.ServiceContext
	data   *types.MessageReceiveBase
	sType  string
	logic  T
}

func DO[T types.SipReceiveHandleLogic[T]](
	Type string,
	svcCtx *types.ServiceContext,
	req sip.Request,
	tx sip.ServerTransaction,
	data *types.MessageReceiveBase,
	logic T,
) {
	var h = &handler[T]{
		svcCtx: svcCtx,
		req:    req,
		tx:     tx,
		logic:  logic,
		data:   data,
		sType:  Type,
	}
	h.run()
}

func (h handler[T]) respond(resp sip.Response) error {
	if h.tx != nil {
		if to, ok := resp.To(); ok {
			resp.ReplaceHeaders("To", []sip.Header{
				&sip.ToHeader{
					Address: to.Address,
					Params: sip.NewParams().Add(
						"tag",
						sip.String{Str: functions.RandWithString("0123456789", 9)},
					),
				},
			})
		}
		if h.svcCtx.Config.UseSipPrintLog {
			SipLog(h.svcCtx.Config.UseSipLogToFile, h.svcCtx.Config.SipLogPath, types.BroadcastTypeSipResponse, h.sType, h.req, resp.String())
		}
		if err := h.tx.Respond(resp); err != nil {
			return err
		}
	}

	return nil
}

func (h handler[T]) run() {
	data, err := ParseToRequest(h.req)
	if err != nil {
		if err := h.respond(sip.NewResponseFromRequest("", h.req, http.StatusBadRequest, "Request parse failed", "")); err != nil {
			functions.LogError("Respond err:", err.Error())
		}
		return
	}

	var content = strings.TrimSuffix(h.req.String(), "\n")
	h.svcCtx.SipLog <- &types.SipLogItem{
		Content: content,
		Type:    types.BroadcastTypeSipReceive,
	}
	if h.svcCtx.Config.UseSipPrintLog {
		SipLog(h.svcCtx.Config.UseSipLogToFile, h.svcCtx.Config.SipLogPath, types.BroadcastTypeSipReceive, h.sType, h.req, content)
	}

	var setting = rule.NewConfig(h.svcCtx.Config, h.svcCtx.Setting).Conv()
	if functions.Contains(data.ID, strings.Split(setting.Content().BanIp, "\n")) {
		_ = h.respond(sip.NewResponseFromRequest("", h.req, types.StatusForbidden, "Forbidden", ""))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(h.svcCtx.Config.Timeout)*time.Millisecond)
	defer cancel()

	var res = h.logic.New(ctx, h.svcCtx, data, h.tx).DO()
	if res != nil && res.Error != nil {
		if h.data == nil {
			h.data, _ = NewParser[types.MessageReceiveBase]().ToData(h.req)
		}

		if h.data != nil {
			functions.LogError(fmt.Sprintf("请求失败, cmdType: %s err: %s", h.data.CmdType, res.Error.Error()))
		} else {
			functions.LogError("请求失败, err: ", res.Error.Error())
		}

		if res.Code == types.StatusPreconditionFailed {
			// h.forbidden()
			return
		}

		var code = res.Code
		if code == 0 {
			code = types.StatusBadRequest
		}

		if code == types.StatusUnauthorized {
			h.unauthorized()
			return
		}
		if code == types.StatusBadRequest {
			h.badRequest()
			return
		}
		if code == types.StatusForbidden {
			h.forbidden()
			return
		}

		if res.Ignore {
			return
		}

		var resp = sip.NewResponseFromRequest("", h.req, code, "", "")
		if res.BeforeResponse != nil {
			resp = res.BeforeResponse(resp)
		}

		if h.tx != nil {
			if err := h.respond(resp); err != nil {
				functions.LogError("Respond err:", err.Error())
			}
		}

		return
	}

	if res != nil && res.Ignore {
		return
	}

	h.success(res)
}

func (h handler[T]) success(res *types.Response) {
	var resp sip.Response
	if res == nil {
		resp = sip.NewResponseFromRequest("", h.req, http.StatusOK, "OK", "")
	} else {
		resp = sip.NewResponseFromRequest("", h.req, http.StatusOK, "OK", res.Data)
		if res.BeforeResponse != nil {
			resp = res.BeforeResponse(resp)
		}
	}

	resp.AppendHeader(
		&sip.GenericHeader{
			HeaderName: "Date",
			Contents:   time.Now().Format("2006-01-02T15:04:05.000"),
		},
	)
	if err := h.respond(resp); err != nil {
		functions.LogError("Respond err:", err.Error())
	}
}

func (h handler[T]) unauthorized() {
	var response = sip.NewResponseFromRequest("", h.req, types.StatusUnauthorized, "Unauthorized", "")
	response.AppendHeader(&sip.GenericHeader{
		HeaderName: HeaderWWWAuthenticateKey,
		Contents: fmt.Sprintf(
			`Digest realm="%s",algorithm=%s,nonce="%s"`,
			h.svcCtx.Config.Sip.ID[0:10],
			"MD5",
			h.randString("0123456789", 32),
		),
	})

	if err := h.respond(response); err != nil {
		functions.LogError("Respond err:", err.Error())
	}
}

func (h handler[T]) forbidden() {
	var resp = sip.NewResponseFromRequest("", h.req, types.StatusForbidden, "Forbidden", "")
	resp.AppendHeader(
		&sip.GenericHeader{
			HeaderName: "Date",
			Contents:   time.Now().Format("2006-01-02T15:04:05.000"),
		},
	)
	if err := h.respond(resp); err != nil {
		functions.LogError("Respond err:", err.Error())
	}
}

func (h handler[T]) badRequest() {
	var resp = sip.NewResponseFromRequest("", h.req, types.StatusBadRequest, "BadRequest", "")
	resp.AppendHeader(
		&sip.GenericHeader{
			HeaderName: "Date",
			Contents:   time.Now().Format("2006-01-02T15:04:05.000"),
		},
	)
	if err := h.respond(resp); err != nil {
		functions.LogError("Respond err:", err.Error())
	}
}

func (h handler[T]) randString(src string, n int) string {
	var (
		output = make([]byte, n)
		srcLen = len(src)
	)
	for key := range output {
		output[key] = src[rand.Intn(srcLen)]
	}

	return string(output)
}
