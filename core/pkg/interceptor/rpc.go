// @Title        interceptor
// @Description  rpc
// @Create       yirl 2025/3/21 11:35

package interceptor

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"skeyevss/core/pkg/functions"
)

const (
	RpcReqTokenKey       = "x-token"
	RpcReqAdminKey       = "x-admin-id"
	RpcReqAdminSuperKey  = "x-admin-super"
	RpcReqAdminDepIdsKey = "x-admin-department-ids"
	RpcReqTimestampKey   = "x-timestamp"

	RpcReqCtxLicenseKey = "x-license"

	aesKey = "ibl4sSDZi4up9MBT"
)

// RPCAuthSenderType rpc客户端链接参数类型
type RPCAuthSenderType struct {
	CKey, // client key
	SKey string // service key
}

// rpc 错误创建 --------------------------------------------

func MakeError(code codes.Code, err error) *Error {
	return &Error{Code: code, Err: err}
}

// --------------------------------------------
// --------------------------------------------
// ---------------------------------- rpc rpc业务代码错误类型
// --------------------------------------------
// --------------------------------------------

type Error struct {
	Code codes.Code
	Err  error
}

func (e *Error) toErr() error {
	return status.Errorf(e.Code, e.Err.Error())
}

// --------------------------------------------
// --------------------------------------------
// -------------------------- rpc 请求拦截 验证器
// --------------------------------------------
// --------------------------------------------

type Interceptor struct {
}

func New() *Interceptor {
	return new(Interceptor)
}

// https://golang2.eddycjy.com/posts/ch3/08-grpc-interceptor/

// Sev rpc服务端拦截器验证
func (*Interceptor) Sev(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
	call func(ctx context.Context, md metadata.MD) (context.Context, *Error),
) (interface{}, error) {
	// 调用前验证
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "auth failed, missed metadata")
	}

	var err *Error
	ctx, err = call(ctx, md)
	if err != nil {
		return nil, err.toErr()
	}

	// 调用
	return handler(ctx, req)
}

// Client rpc客户端发送拦截器
func (*Interceptor) Client(
	call func(md metadata.MD) (metadata.MD, error),
	ctx context.Context,
	method string,
	req,
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	var (
		md  = metadata.New(map[string]string{RpcReqTimestampKey: strconv.FormatInt(time.Now().UnixMilli(), 10)})
		err error
	)
	md, err = call(md)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "metadata make failed")
	}

	return invoker(metadata.NewOutgoingContext(ctx, md), method, req, reply, cc, opts...)
}

// --------------------------------------------
// --------------------------------------------
// ---------------------------------- rpc token
// --------------------------------------------
// --------------------------------------------

type RPCToken struct {
	CKey, // client key
	SKey, // service key
	Timestamp string
}

func NewRPCToken(sKey string) *RPCToken {
	return &RPCToken{
		SKey: sKey,
	}
}

// Make 生成token
func (r *RPCToken) Make(cKey string, md metadata.MD) (string, error) {
	var timestamp = md.Get(RpcReqTimestampKey)
	b, err := functions.JSONMarshal(&RPCToken{
		CKey:      cKey,
		SKey:      r.SKey,
		Timestamp: strings.Join(timestamp, ""),
	})
	if err != nil {
		return "", err
	}

	encrypt, err := functions.NewCrypto([]byte(aesKey)).Encrypt(b)
	if err != nil {
		return "", err
	}

	return encrypt, nil
}

// Parse 解析token
func (r *RPCToken) Parse(ctx context.Context, md metadata.MD) (context.Context, error) {
	var (
		timestamp = strings.Join(md.Get(RpcReqTimestampKey), "")
		token     = strings.Join(md.Get(RpcReqTokenKey), "")
		adminId   = md.Get(RpcReqAdminKey)
		depIds    = md.Get(RpcReqAdminDepIdsKey)
		super     = md.Get(RpcReqAdminSuperKey)
	)
	if timestamp == "" || token == "" {
		return ctx, fmt.Errorf("metadata parameters [%s] [%s] is empty", RpcReqTimestampKey, RpcReqTimestampKey)
	}

	b, err := functions.NewCrypto([]byte(aesKey)).Decrypt(token)
	if err != nil {
		return ctx, err
	}

	var data RPCToken
	if err := functions.JSONUnmarshal([]byte(b), &data); err != nil {
		return ctx, err
	}

	if data.Timestamp != timestamp {
		return ctx, fmt.Errorf("invalid token")
	}

	if data.SKey != r.SKey {
		return ctx, fmt.Errorf("invalid skey")
	}

	if len(adminId) <= 0 {
		return context.WithValue(ctx, RpcReqAdminKey, "0"), nil
	}

	ctx = context.WithValue(ctx, RpcReqAdminKey, adminId[0])
	if len(depIds) > 0 {
		ctx = context.WithValue(ctx, RpcReqAdminDepIdsKey, depIds[0])
	}

	if len(super) > 0 {
		ctx = context.WithValue(ctx, RpcReqAdminSuperKey, super[0])
	}

	return ctx, nil
}

func GetAdminId(ctx context.Context) uint64 {
	v, ok := ctx.Value(RpcReqAdminKey).(string)
	if !ok {
		return 0
	}

	id, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0
	}

	return id
}

func GetAdminSuper(ctx context.Context) uint64 {
	v, ok := ctx.Value(RpcReqAdminSuperKey).(string)
	if !ok {
		return 0
	}

	id, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0
	}

	return id
}

func GetAdminDepIds(ctx context.Context) []uint64 {
	v, ok := ctx.Value(RpcReqAdminDepIdsKey).(string)
	if !ok {
		return nil
	}

	var ids []uint64
	if err := functions.JSONUnmarshal([]byte(v), &ids); err != nil {
		functions.LogError("unmarshal admin dep ids failed", err)
		return nil
	}

	return ids
}
