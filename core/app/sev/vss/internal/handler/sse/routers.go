// @Title        routers
// @Description  main
// @Create       yiyiyi 2025/8/12 14:20

package sse

import (
	"context"
	"fmt"
	"net/http"

	v "github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"

	"skeyevss/core/app/sev/vss/internal/logic/sse"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/response"
)

func sseHandler[Logic any, Req types.SSERequestType](r *http.Request, handler types.SSEHandleLogic[Logic, Req], req Req, messageChan chan *types.SSEResponse) {
	if err := schema.NewDecoder().Decode(req, r.URL.Query()); err != nil {
		messageChan <- &types.SSEResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("url参数解析错误, err: %s", err)), localization.M0001),
		}
		return
	}

	if err := v.New().Struct(req); err != nil {
		messageChan <- &types.SSEResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("参数校验失败, err: %s", err)), localization.M0001),
		}
		return
	}

	go handler.DO(req)
}

func RegisterRouter(ctx context.Context, r *http.Request, svcCtx *types.ServiceContext, messageChan chan *types.SSEResponse) {
	switch r.URL.Query().Get("type") {
	case sse.VDeviceDiagnoses.GetType():
		sseHandler(r, sse.VDeviceDiagnoses.New(ctx, svcCtx, messageChan), new(sse.SSEDeviceDiagnosesReq), messageChan)

	case sse.VFileDownloadLogic.GetType():
		sseHandler(r, sse.VFileDownloadLogic.New(ctx, svcCtx, messageChan), new(sse.SSEFileDownloadReq), messageChan)

	case sse.VSevState.GetType():
		sseHandler(r, sse.VSevState.New(ctx, svcCtx, messageChan), new(sse.SSESevStateReq), messageChan)

	case sse.VChannelDiagnoses.GetType():
		sseHandler(r, sse.VChannelDiagnoses.New(ctx, svcCtx, messageChan), new(sse.SSEChannelDiagnosesReq), messageChan)

	case sse.VDeviceOnlineStates.GetType():
		sseHandler(r, sse.VDeviceOnlineStates.New(ctx, svcCtx, messageChan), new(sse.SSEDeviceOnlineStatesReq), messageChan)

	case sse.VSipLogs.GetType():
		sse.VSipLogs.New(ctx, svcCtx, messageChan).DO()

	default:
		messageChan <- &types.SSEResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("type 不能为空"), localization.M0001),
		}
	}
}
