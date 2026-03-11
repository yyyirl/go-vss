package ws

import (
	"errors"

	"skeyevss/core/app/sev/vss/internal/pkg"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
)

type response struct {
	svcCtx *types.ServiceContext
	client *types.WSClient
}

func newResponse(svcCtx *types.ServiceContext, client *types.WSClient) *response {
	return &response{svcCtx, client}
}

// 响应当前客户端
func (r *response) current(msg *types.WSResponseMessage) error {
	return r.client.WebsocketConn.WriteMessage(msg.MessageType, r.data(msg, r.svcCtx.Config.Host))
}

// 向其他人发送消息
func (r *response) somebody(msg *types.WSResponseMessage, userid uint64) error {
	// 本机节点缓存
	if client := r.svcCtx.WSClientCache.Row(pkg.NewUtils(nil).MakeClientId(r.client.ConnType, userid)); client != nil {
		return client.WebsocketConn.WriteMessage(msg.MessageType, r.data(msg, r.svcCtx.Config.Host))
	}

	return errors.New("指定用户消息发送失败")
}

func (r *response) data(response *types.WSResponseMessage, host string) []byte {
	return r.responseMarshal(response, host)
}

func (r *response) responseMarshal(rep *types.WSResponseMessage, _ string) []byte {
	data, _ := functions.JSONMarshal(rep.ToRespMap())

	return data
}
