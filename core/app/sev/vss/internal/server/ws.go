package server

import (
	"fmt"
	"net/http"

	"skeyevss/core/app/sev/vss/internal/handler/ws"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
)

type WSSev struct {
	svcCtx *types.ServiceContext
}

func NewWSSev(svcCtx *types.ServiceContext) *WSSev {
	return &WSSev{
		svcCtx: svcCtx,
	}
}

func (l *WSSev) Start() {
	var addr = fmt.Sprintf("%s:%d", l.svcCtx.Config.Host, l.svcCtx.Config.WS.Port)
	functions.PrintStyle("blue", "Websocket Listen on: ", addr)

	// 定时器 链接检测
	ws.NewInterval(l.svcCtx).Do()
	// 消息处理器
	go ws.NewProc(l.svcCtx).Receiver()
	// handler
	http.HandleFunc("/", ws.NewWSSev(l.svcCtx).Do)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
