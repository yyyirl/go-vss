// @Title        middleware
// @Description  main
// @Create       yirl 2025/3/21 12:44

package middleware

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"skeyevss/core/app/sev/db/internal/config"
	"skeyevss/core/pkg/interceptor"
)

func Sev(conf *config.Config, _ string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		return interceptor.New().Sev(ctx, req, info, handler, func(ctx context.Context, md metadata.MD) (context.Context, *interceptor.Error) {
			// 校验服务鉴权
			ctx, err := interceptor.NewRPCToken(conf.SevBase.Keys.DB).Parse(ctx, md)
			if err != nil {
				return ctx, interceptor.MakeError(codes.PermissionDenied, fmt.Errorf("校验失败[1], %s", err))
			}

			return context.WithValue(ctx, interceptor.RpcReqCtxLicenseKey, ""), nil
			// // 验证激活信息
			// data, err := license.NewVerify().Verification(conf.ActivateCodePath, buildTime)
			// if err != nil {
			// 	// 没有授权信息
			// 	return context.WithValue(ctx, interceptor.RpcReqCtxLicenseKey, ""), nil
			// 	// return ctx, interceptor.MakeError(codes.PermissionDenied, fmt.Errorf("授权信息校验失败, %s", err))
			// }
			//
			// return context.WithValue(ctx, interceptor.RpcReqCtxLicenseKey, data.ToString()), nil
		})
	}
}
