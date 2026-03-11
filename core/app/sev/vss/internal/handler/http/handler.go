package http

import (
	"context"
	"net/url"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zeromicro/go-zero/rest/httpx"

	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/constants"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

type xHandler[Req any, Logic types.HttpRHandleLogic[Logic, Req]] struct {
	svcCtx *types.ServiceContext
	logic  Logic
}

func newHandlerWithParams[Req any, Logic types.HttpRHandleLogic[Logic, Req]](svcCtx *types.ServiceContext, logic Logic) gin.HandlerFunc {
	var l = &xHandler[Req, Logic]{
		svcCtx: svcCtx,
		logic:  logic,
	}

	return l.do
}

func (p *xHandler[Req, Logic]) do(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			functions.LogError("vss http recover, err: ", err, string(debug.Stack()))
		}
	}()

	var (
		ctx = c.Request.Context()
		req Req
	)
	if err := httpx.Parse(c.Request, &req); err != nil {
		response.New().RequestError(ctx, c.Writer, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001))
		return
	}

	// 获取当前请求ip
	ctx = context.WithValue(ctx, constants.HEADER_IP, c.ClientIP())
	ctx = context.WithValue(ctx, constants.CTX_VSS_IS_INTERNAL_REQ, false)
	if parsedURL, _ := url.Parse(c.GetHeader("Referer")); parsedURL != nil {
		var host = strings.Split(parsedURL.Host, ":")[0]
		ctx = context.WithValue(
			ctx,
			constants.CTX_VSS_IS_INTERNAL_REQ,
			functions.Contains(
				host,
				[]string{
					"127.0.0.1", "::1", "localhost",
				},
			) || host == p.svcCtx.Config.InternalIp,
		)
	}

	toResp(c, p.logic.New(ctx, c, p.svcCtx).DO(req))
}

type handler[Logic types.HttpEHandleLogic[Logic]] struct {
	svcCtx *types.ServiceContext
	logic  Logic
}

func newHandler[Logic types.HttpEHandleLogic[Logic]](svcCtx *types.ServiceContext, logic Logic) gin.HandlerFunc {
	var l = &handler[Logic]{
		svcCtx: svcCtx,
		logic:  logic,
	}

	return l.do
}

func (p *handler[Logic]) do(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			functions.LogError("vss http recover, err:", err, string(debug.Stack()))
		}
	}()

	var ctx = context.WithValue(c.Request.Context(), constants.HEADER_IP, c.ClientIP())
	ctx = context.WithValue(ctx, constants.CTX_VSS_IS_INTERNAL_REQ, false)
	if parsedURL, _ := url.Parse(c.GetHeader("Referer")); parsedURL != nil {
		var host = strings.Split(parsedURL.Host, ":")[0]
		ctx = context.WithValue(
			ctx,
			constants.CTX_VSS_IS_INTERNAL_REQ,
			functions.Contains(
				host,
				[]string{
					"127.0.0.1", "::1", "localhost",
				},
			) || host == p.svcCtx.Config.InternalIp,
		)
	}

	toResp(c, p.logic.New(ctx, c, p.svcCtx).DO())
}

func toResp(c *gin.Context, resp *types.HttpResponse) {
	var ctx = c.Request.Context()
	if resp == nil {
		response.New().Success(ctx, c.Writer, nil)
		return
	}

	if resp.Err != nil {
		response.New().RequestError(ctx, c.Writer, response.MakeError(response.NewHttpRespMessage().Str(resp.Err.Error), resp.Err.Message))
		return
	}

	if resp.Data != nil {
		response.New().Success(ctx, c.Writer, resp.Data)
		return
	}

	response.New().Success(ctx, c.Writer, nil)
}
