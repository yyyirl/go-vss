package ws

import (
	"context"
	"time"

	validator "github.com/go-playground/validator/v10"

	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/tps"
)

// router --------------------------------------------------------
type router struct {
	svcCtx *types.ServiceContext
	client *types.WSClient
}

func newRouters(svcCtx *types.ServiceContext, client *types.WSClient) *router {
	return &router{svcCtx: svcCtx, client: client}
}

func (r router) requestParse(req *types.WSRequestContent, data interface{}) *types.WSResponse {
	// 解析data
	if err := functions.ConvInterface(req.Data, &data); err != nil {
		return &types.WSResponse{
			Errors: &tps.XError{Message: err.Error()},
		}
	}

	if err := validator.New().Struct(data); err != nil {
		return &types.WSResponse{
			Errors: &tps.XError{Message: err.Error()},
		}
	}

	return nil
}

func (r router) dispatch(req *types.WSRequestContent) *types.WSResponseMessage {
	r.client.ActiveTime = functions.NewTimer().Now()
	var resp = new(types.WSResponseMessage)
	resp.WSResponse = new(types.WSResponse)
	resp.MessageType = req.MessageType

	item, ok := routers[req.Type]
	if ok {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.svcCtx.Config.WS.ReqTimeout)*time.Millisecond)
		defer cancel()

		resp.WSResponse = item.handler(r.svcCtx, &types.WSHandlerCallParams{
			Ctx:          ctx,
			Client:       r.client,
			Req:          req,
			RequestParse: r.requestParse,
		})
	}

	if resp.WSResponse != nil {
		resp.Type = req.Type
	}

	return resp
}

// 广播 --------------------------------------------------------

type broadcaster struct {
	svcCtx *types.ServiceContext
}

func newBroadcaster(svcCtx *types.ServiceContext) *broadcaster {
	return &broadcaster{svcCtx: svcCtx}
}

func (r broadcaster) dispatch(req *types.BroadcastMessageItem) error {
	item, ok := broadcasters[req.Type]
	if ok {
		return item.handler(r.svcCtx, req.Data)
	}

	return nil
}
