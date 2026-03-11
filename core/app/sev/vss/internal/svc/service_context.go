// @Title        middleware
// @Description  main
// @Create       yirl 2025/6/9 14:00

package svc

import (
	"time"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	backendClient "skeyevss/core/app/sev/db/client/backendservice"
	configClient "skeyevss/core/app/sev/db/client/configservice"
	deviceClient "skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/vss/internal/config"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/common/client"
	"skeyevss/core/pkg/audio"
	"skeyevss/core/pkg/broadcast"
	"skeyevss/core/pkg/categories"
	"skeyevss/core/pkg/functions/download"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/pkg/set"
	"skeyevss/core/pkg/xmap"
	"skeyevss/core/repositories/models/cascade"
	"skeyevss/core/repositories/models/dictionaries"
	"skeyevss/core/repositories/models/settings"
	"skeyevss/core/repositories/redis"
)

func NewServiceContext(c config.Config) *types.ServiceContext {
	var (
		rpcInterceptor   = client.NewRpcClientInterceptor(c.RpcInterceptor)
		rpcClientOptions = rpcInterceptor.Options(map[client.OptionsKey]zrpc.ClientOption{
			client.OptionsRetryKey: zrpc.WithUnaryClientInterceptor(
				rpcInterceptor.RetryInterceptor(
					c.RpcInterceptor.RpcCallerRetryMax,
					time.Duration(c.RpcInterceptor.RpcCallerRetryWaitInterval)*time.Millisecond,
				),
			),
			client.OptionsKeepaliveKey: zrpc.WithDialOption(
				grpc.WithKeepaliveParams(keepalive.ClientParameters{
					Time:                time.Duration(c.RpcInterceptor.RpcKeepaliveTime) * time.Second,
					Timeout:             time.Duration(c.RpcInterceptor.RpcKeepaliveTimeout) * time.Second,
					PermitWithoutStream: c.RpcInterceptor.RpcKeepalivePermitWithoutStream,
				}),
			),
			client.OptionsApi2RpcKey: zrpc.WithUnaryClientInterceptor(
				rpcInterceptor.Api2DBRpc(
					&interceptor.RPCAuthSenderType{
						SKey: c.SevBase.Keys.DB,
						CKey: c.SevBase.Keys.BackendApi,
					},
				),
			),
		})
	)

	return &types.ServiceContext{
		Config: c,
		RpcClients: &client.GRPCClients{
			Backend: backendClient.NewBackendService(zrpc.MustNewClient(c.DBGrpc, rpcClientOptions...)),
			Config:  configClient.NewConfigService(zrpc.MustNewClient(c.DBGrpc, rpcClientOptions...)),
			Device:  deviceClient.NewDeviceService(zrpc.MustNewClient(c.DBGrpc, rpcClientOptions...)),
		},
		RedisClient: redis.New(c.Mode, c.Log.Encoding, c.Redis, c.Log),
		Broadcast:   broadcast.NewBroadcast(100),

		WSProc: &types.WSProc{
			ReceiveMessageChan:  make(chan *types.WSMessageReceiveItem, 100),
			ResponseMessageChan: make(chan *types.WSResponseMessageItem, 100),
			BroadcastChan:       make(chan *types.BroadcastMessageItem, 100),
			CloseChan:           make(chan *types.WSCloseChanItem, 100),
		},
		WSClientCache:     types.NewWSClientsCache(),
		WSTalkUsageStatus: xmap.New[string, string](100),
		TalkSipData:       xmap.New[string, *audio.TalkSessionItem](100),
		TalkSipSendStatus: set.New[string](100),

		SipSendCatalog:           make(chan *types.Request, 100),
		SipSendDeviceInfo:        make(chan *types.Request, 100),
		SipSendVideoLiveInvite:   make(chan *types.SipVideoLiveInviteMessage, 100),
		SipSendTalkInvite:        make(chan *types.SipTalkInviteMessage, 100),
		SipSendBye:               make(chan *types.SipByeMessage, 100),
		SipSendDeviceControl:     make(chan *types.DeviceControlReq, 100),
		SipSendQueryPresetPoints: make(chan *types.SipSendQueryPresetPointsReq, 100),
		SipSendSetPresetPoints:   make(chan *types.SipSendSetPresetPointsReq, 100),
		SipSendQueryVideoRecords: make(chan *types.QueryVideoRecordsReq, 100),
		SipSendSubscription:      make(chan *types.SubscriptionReq, 100),
		SipSendBroadcast:         make(chan *types.BroadcastReq, 100),
		SipSendTalk:              make(chan *types.GBSSipSendTalk, 100),
		SipLog:                   make(chan *types.SipLogItem, 100),

		SipCatalogLoop:   make(chan *types.SipCatalogLoopReq, 100),
		SipHeartbeatLoop: make(chan *types.SipHeartbeatLoopReq, 100),
		SetDeviceOnline:  make(chan *types.DCOnlineReq, 100),

		SipGBSSNMap:                xmap.New[string, uint32](1000),
		SipGBCSNMap:                xmap.New[string, uint32](1000),
		SipCatalogLoopMap:          xmap.New[string, *types.SipCatalogLoopReq](1000),
		SipHeartbeatLoopMap:        xmap.New[string, *types.SipHeartbeatLoopReq](1000),
		DeviceOnlineStateUpdateMap: xmap.New[string, *types.DCOnlineReq](1000),
		AckRequestMap:              xmap.New[string, *types.SendSipRequest](1000),
		FetchDeviceVideoState:      set.New[string](100),

		PubStreamExistsState: set.New[string](1000),
		InviteRequestState:   set.New[string](1000),

		// PlaybackControlMap:        xmap.New[string, *types.PlaybackControlItem](1000),
		SipMessagePresetPointsMap: xmap.New[string, *types.SipMessageQueryPresetPointsResp](1000),
		SipMessageVideoRecordMap:  xmap.New[string, *types.SipMessageVideoRecords](1000),

		DictionaryMap: make(map[string]*categories.Item[int, *dictionaries.Item]),
		Setting: &settings.Item{
			Content: new(settings.Content),
		},

		GBCRegisterChan:  make(chan *types.GBCRegisterChanItem, 100),
		GBCKeepaliveChan: make(chan *cascade.Item, 100),

		GBCInviteReqMaps:      xmap.New[string, *types.SipGBCInviteReqItem](1000),
		GBCRecordInfoSendMaps: xmap.New[string, *types.GBCRecordInfoItem](100),

		CascadeRegister:          xmap.New[uint64, *types.CascadeRegisterItem](1000),
		CascadeKeepaliveCounter:  xmap.New[uint64, uint64](1000),
		CascadeRegisterExecuting: xmap.New[string, bool](100),

		DownloadManager: download.GetManager(),
	}
}
