package svc

import (
	"time"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"skeyevss/core/app/sev/cron/internal/config"
	"skeyevss/core/app/sev/cron/internal/types"
	backendClient "skeyevss/core/app/sev/db/client/backendservice"
	configClient "skeyevss/core/app/sev/db/client/configservice"
	deviceClient "skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/common/client"
	"skeyevss/core/pkg/interceptor"
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
		Config:      c,
		RedisClient: redis.New(c.Mode, c.Log.Encoding, c.Redis, c.Log),
		RpcClients: &client.GRPCClients{
			Backend: backendClient.NewBackendService(zrpc.MustNewClient(c.DBGrpc, rpcClientOptions...)),
			Config:  configClient.NewConfigService(zrpc.MustNewClient(c.DBGrpc, rpcClientOptions...)),
			Device:  deviceClient.NewDeviceService(zrpc.MustNewClient(c.DBGrpc, rpcClientOptions...)),
		},

		Data: new(types.SvcData),
	}
}
