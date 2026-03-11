// @Title        response
// @Description  rpc
// @Create       yirl 2025/3/21 16:29

package response

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"

	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
)

// ------------------------------- rpc函数返回错误信息

const rpcErrPrefix = "\nrpc-caller: "

func NewMakeRpcRetErr(err error, skip int) error {
	if err == nil {
		return nil
	}

	return errors.New(err.Error() + rpcErrPrefix + functions.CallerFile(skip))
}

// ------------------------------- rpc服务端响应

type RpcResp[T proto.Message] struct{}

func NewRpcResp[T proto.Message]() *RpcResp[T] {
	return new(RpcResp[T])
}

func (*RpcResp[T]) XMake(data proto.Message, call func(data *anypb.Any) T) (T, error) {
	res, err := proto.Marshal(data)
	if err != nil {
		var resp T
		return resp, err
	}

	return call(&anypb.Any{Value: res}), nil
}

func (*RpcResp[T]) Make(data interface{}, skip int, call func(data []byte) T) (T, error) {
	b, err := functions.JSONMarshal(data)
	if err != nil {
		var resp T
		return resp, NewMakeRpcRetErr(err, skip)
	}

	return call(b), nil
}

// ------------------------------- rpc客户端响应解析为http error, any data

type RpcToHttpXResp[T, V proto.Message] struct{}

func NewRpcToHttpXResp[T, V proto.Message]() *RpcToHttpXResp[T, V] {
	return new(RpcToHttpXResp[T, V])
}

func (*RpcToHttpXResp[T, V]) XParse(value V, call func() (T, error)) *HttpErr {
	res, err := call()
	if err != nil {
		if v, ok := status.FromError(err); ok {
			switch v.Code() {
			case codes.InvalidArgument:
				return MakeError(NewHttpRespMessage().Err(err), localization.MR1004)

			case codes.DeadlineExceeded:
				return MakeError(NewHttpRespMessage().Err(err), localization.MR1005)

			case codes.PermissionDenied:
				return MakeError(NewHttpRespMessage().Err(err), localization.MR1006)

			default:
				return MakeError(NewHttpRespMessage().Err(err), localization.Make(v.Message()))
				// return MakeError(NewHttpRespMessage().Err(err), localization.MR1003)
			}
		}

		return MakeError(NewHttpRespMessage().Err(err), localization.MR1003)
	}

	data, err := functions.GetFieldValue(res, "Data")
	if err != nil {
		return MakeError(NewHttpRespMessage().Err(err), localization.MR1001)
	}

	anyData, ok := data.(*anypb.Any)
	if !ok {
		return MakeError(NewHttpRespMessage().Str(fmt.Sprintf("type Error, data is not anypb.Any, input %T", data)), localization.MR1000)
	}

	if err := proto.Unmarshal(anyData.Value, value); err != nil {
		return MakeError(NewHttpRespMessage().Err(err), localization.MR1002)
	}

	return nil
}

// ------------------------------- rpc客户端响应解析为http error, bytes data

type RpcToHttpRespType[T proto.Message, V any] struct {
	Res  T
	Data V
}

type RpcToHttpResp[T proto.Message, V any] struct{}

func NewRpcToHttpResp[T proto.Message, V any]() *RpcToHttpResp[T, V] {
	return new(RpcToHttpResp[T, V])
}

func (*RpcToHttpResp[T, V]) Parse(call func() (T, error)) (*RpcToHttpRespType[T, V], *HttpErr) {
	res, err := call()
	if err != nil {
		if v, ok := status.FromError(err); ok {
			switch v.Code() {
			case codes.InvalidArgument:
				return nil, MakeError(NewHttpRespMessage().Err(errors.New(err.Error()+rpcErrPrefix+functions.CallerFile(2))), localization.MR1004)

			case codes.DeadlineExceeded:
				return nil, MakeError(NewHttpRespMessage().Err(errors.New(err.Error()+rpcErrPrefix+functions.CallerFile(2))), localization.MR1005)

			case codes.PermissionDenied:
				return nil, MakeError(NewHttpRespMessage().Err(errors.New(err.Error()+rpcErrPrefix+functions.CallerFile(2))), localization.MR1006)

			default:
				return nil, MakeError(NewHttpRespMessage().Err(errors.New(err.Error()+rpcErrPrefix+functions.CallerFile(2))), localization.Make(v.Message()))
				// return nil, MakeError(NewHttpRespMessage().Err(errors.New(err.Error() + rpcErrPrefix + functions.CallerFile(2))), localization.MR1003)
			}
		}

		return nil, MakeError(NewHttpRespMessage().Err(errors.New(err.Error()+rpcErrPrefix+functions.CallerFile(2))), localization.MR1003)
	}

	data, err := functions.GetFieldValue(res, "Data")
	if err != nil {
		return nil, MakeError(NewHttpRespMessage().Err(errors.New(err.Error()+rpcErrPrefix+functions.CallerFile(2))), localization.MR1001)
	}

	b, ok := data.([]byte)
	if !ok {
		return nil, MakeError(NewHttpRespMessage().Str(fmt.Sprintf("type Error, data is not anypb.Any, input %T caller %s", data, functions.CallerFile(2))), localization.MR1000)
	}

	var value V
	if ok, vt := functions.IsSimpleType(value); ok {
		if vt.Kind() == reflect.Bool && len(b) > 0 {
			v, _ := strconv.ParseBool(string(b))
			return &RpcToHttpRespType[T, V]{
				Res:  res,
				Data: any(v).(V),
			}, nil
		}

		_ = functions.JSONUnmarshal(b, &value)
		return &RpcToHttpRespType[T, V]{
			Res:  res,
			Data: value,
		}, nil
	}

	if len(b) <= 0 {
		return &RpcToHttpRespType[T, V]{
			Res:  res,
			Data: value,
		}, nil
	}

	if err := functions.JSONUnmarshal(b, &value); err != nil {
		return nil, MakeError(NewHttpRespMessage().Err(errors.New(err.Error()+rpcErrPrefix+functions.CallerFile(2))), localization.MR1002)
	}

	return &RpcToHttpRespType[T, V]{
		Res:  res,
		Data: value,
	}, nil
}
