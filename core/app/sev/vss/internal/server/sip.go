package server

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/ghettovoice/gosip"

	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
)

// server ------------------------------------------------------
type (
	SipServer struct {
		networkType string
		svcCtx      *types.ServiceContext
	}

	SipNetworkType string
)

const (
	SipUDP SipNetworkType = "udp"
	SipTCP SipNetworkType = "tcp"
)

func NewSipSev(svcCtx *types.ServiceContext) *SipServer {
	return &SipServer{
		svcCtx: svcCtx,
	}
}

func (s *SipServer) SipGbsServer(networkType SipNetworkType, handlers types.HType) {
	s.svcCtx.InitFetchDataState.Wait()

	var (
		sipSvr = gosip.NewServer(gosip.ServerConfig{Host: s.svcCtx.Config.InternalIp}, nil, nil, NewLogger())
		addr   = fmt.Sprintf("%s:%d", s.svcCtx.Config.Host, s.svcCtx.Config.Sip.Port)
	)
	for key, item := range handlers {
		if err := sipSvr.OnRequest(key, item); err != nil {
			functions.LogError(fmt.Sprintf("Sip GBS Request [%S] err: %s", key, err.Error()))
		}
	}

	if err := sipSvr.Listen(string(networkType), addr); err != nil {
		panic(err)
	}

	if networkType == SipTCP {
		s.svcCtx.GBSTCPSev = &sipSvr
	} else {
		s.svcCtx.GBSUDPSev = &sipSvr
	}

	functions.PrintStyle("blue", "SIP GBS Listen on: ", addr)
}

func (s *SipServer) SipGbcServer(networkType SipNetworkType, handlers types.HType) {
	s.svcCtx.InitFetchDataState.Wait()

	var (
		sipSvr = gosip.NewServer(gosip.ServerConfig{Host: s.svcCtx.Config.InternalIp}, nil, nil, NewLogger())
		addr   = fmt.Sprintf("%s:%d", s.svcCtx.Config.Host, s.svcCtx.Config.Sip.CascadeSipPort)
	)
	for key, item := range handlers {
		if err := sipSvr.OnRequest(key, item); err != nil {
			functions.LogError(fmt.Sprintf("Sip GBC Request [%S] err: %s", key, err.Error()))
		}
	}

	if err := sipSvr.Listen(string(networkType), addr); err != nil {
		panic(err)
	}

	if networkType == SipUDP {
		s.svcCtx.GBCUDPSev = &sipSvr
	} else {
		s.svcCtx.GBCTCPSev = &sipSvr
	}

	functions.PrintStyle("blue", "SIP GBC Listen on: ", addr)
}

// server ------------------------------------------------------

// proc --------------------------------------------------------

type SipProc struct {
	svcCtx *types.ServiceContext
}

func NewSipProc(svcCtx *types.ServiceContext) *SipProc {
	return &SipProc{
		svcCtx: svcCtx,
	}
}

func (p *SipProc) DO(options ...types.SipProcLogic) {
	for _, item := range options {
		go item.DO(&types.DOProcLogicParams{
			SvcCtx: p.svcCtx,
			RecoverCall: func(name string) {
				if err := recover(); err != nil {
					functions.LogError(fmt.Sprintf("Sip Interval [%s] Recover [%s] \nStack: %s", name, err, string(debug.Stack())))
					// 防止高频panic
					time.Sleep(1 * time.Second)
					functions.LogInfo("restart interval server ...")
					p.DO(options...)
				}
			},
		})
	}
}

// proc --------------------------------------------------------
