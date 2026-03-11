package ws

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"

	"skeyevss/core/app/sev/vss/internal/pkg"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
)

type Handler struct {
	svcCtx *types.ServiceContext
}

func NewWSSev(svcCtx *types.ServiceContext) *Handler {
	return &Handler{
		svcCtx: svcCtx,
	}
}

func (h *Handler) Do(resp http.ResponseWriter, r *http.Request) {
	// recover
	defer pkg.NewRecover(&h.svcCtx.Config, nil)

	var (
		timer  = functions.NewTimer()
		now    = timer.Now()
		client = new(types.WSClient)

		checkOriginErr error

		subProtocols = websocket.Subprotocols(r)
		upgrade      = websocket.Upgrader{
			ReadBufferSize:  h.svcCtx.Config.WS.ReadBufferMaxSize,
			WriteBufferSize: h.svcCtx.Config.WS.WriteBufferMaxSize,
			Subprotocols:    subProtocols,
			CheckOrigin: func(r *http.Request) bool {
				if len(subProtocols) != 2 {
					checkOriginErr = errors.New("header Sec-Websocket-Protocol 参数错误错误")
					return false
				}

				var (
					connType = subProtocols[0]
					token    = subProtocols[1]
					// xAuthorization = subProtocols[2]
				)
				if connType == "" {
					checkOriginErr = errors.New("header Conn-Type 类型错误")
					return false
				}

				// 链接类型 区分frontend backend
				client.ConnType = connType
				// 链接时间
				client.ActiveTime = now
				client.ConnTime = now
				// 默认分配一个唯一id
				client.ClientId = functions.UniqueId()
				client.Userid = functions.UniqueId()

				data, err := pkg.NewAes(h.svcCtx.Config).ParseUserToken(token)
				if err != nil {
					checkOriginErr = fmt.Errorf("token 无效 %s", err)
					return false
				}

				// 设置连接信息
				client.Validate = true
				client.Userid = strconv.FormatUint(data.ID, 10)
				client.ClientId = pkg.NewUtils(nil).MakeClientId(client.ConnType, data.ID)
				client.Token = token
				// 链接缓存
				h.svcCtx.WSClientCache.Delete(client)
				h.svcCtx.WSClientCache.Add(client.ClientId, client)
				return true
			},
		}
	)

	conn, err := upgrade.Upgrade(resp, r, nil)
	if err != nil {
		functions.LogInfo("链接失败 upgrade err: ", err, "; original error: ", checkOriginErr)
		return
	}

	// 设置client
	client.WebsocketConn = conn
	client.ResponseTo = func(message *types.WSResponseMessage, userid uint64) error {
		return newResponse(h.svcCtx, client).somebody(message, userid)
	}

	go NewProc(h.svcCtx).reader(client)
}
