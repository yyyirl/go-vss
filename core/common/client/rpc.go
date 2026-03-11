// @Title        client
// @Description  rpc
// @Create       yirl 2025/3/21 10:32

package client

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	backendClient "skeyevss/core/app/sev/db/client/backendservice"
	configClient "skeyevss/core/app/sev/db/client/configservice"
	deviceClient "skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/interceptor"
	"skeyevss/core/tps"
)

type OptionsKey = int

const (
	OptionsRetryKey     OptionsKey = 0
	OptionsApi2RpcKey   OptionsKey = 2
	OptionsKeepaliveKey OptionsKey = 3
)

type (
	GRPCClients struct {
		Backend backendClient.BackendService
		Config  configClient.ConfigService
		Device  deviceClient.DeviceService
	}

	RpcClientInterceptor struct {
		conf tps.YamlRpcInterceptorConf
	}
)

func NewRpcClientInterceptor(conf tps.YamlRpcInterceptorConf) *RpcClientInterceptor {
	return &RpcClientInterceptor{conf: conf}
}

func (r *RpcClientInterceptor) Options(options map[OptionsKey]zrpc.ClientOption) []zrpc.ClientOption {
	var list []zrpc.ClientOption
	for key, option := range options {
		if key == OptionsRetryKey {
			if !r.conf.UseRpcCallerRetry {
				continue
			}
		}

		if key == OptionsKeepaliveKey {
			if !r.conf.UseRpcKeepalive {
				continue
			}
		}

		list = append(list, option)
	}

	return list
}

func (*RpcClientInterceptor) isConnectionError(err error) bool {
	return err != nil && (strings.Index(err.Error(), "connection refused") >= 0 ||
		strings.Index(err.Error(), "transport: Error while dialing") >= 0 ||
		strings.Index(err.Error(), "target machine actively refused it") >= 0)
}

func (r *RpcClientInterceptor) RetryInterceptor(maxRetry uint, waitTime time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var (
			err error
			i   uint = 0
		)
		for ; i < maxRetry; i++ {
			err = invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}

			functions.LogInfo("rpc RetryInterceptor, err: ", err.Error())
			if !r.isConnectionError(err) {
				return err
			}

			if i == maxRetry-1 {
				return err
			}

			select {
			case <-time.After(waitTime * time.Duration(i+1)):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return err
	}
}

// Api2DBRpc api服务请求DBRpc 服务拦截器
func (*RpcClientInterceptor) Api2DBRpc(data *interceptor.RPCAuthSenderType) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return interceptor.New().Client(
			func(md metadata.MD) (metadata.MD, error) {
				token, err := interceptor.NewRPCToken(data.SKey).Make(data.CKey, md)
				if err != nil {
					return nil, err
				}

				md.Set(interceptor.RpcReqTokenKey, token)
				md.Set(interceptor.RpcReqAdminKey, strconv.Itoa(int(contextx.GetCtxUserid(ctx))))
				md.Set(interceptor.RpcReqAdminSuperKey, strconv.Itoa(int(contextx.GetSuperState(ctx))))

				var departmentIds = contextx.GetDepartmentIds(ctx)
				if len(departmentIds) > 0 {
					var depIds = "[]"
					b, err := functions.JSONMarshal(departmentIds)
					if err != nil {
						functions.LogError("depIds 序列化失败, err:", err)
					} else {
						depIds = string(b)
					}
					md.Set(interceptor.RpcReqAdminDepIdsKey, depIds)
				}

				return md, nil
			},
			ctx, method, req, reply, cc, invoker, opts...,
		)
	}
}
