package gbs_sip

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ghettovoice/gosip/sip"
	"google.golang.org/protobuf/types/known/structpb"

	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/devices"
)

var _ types.SipReceiveHandleLogic[*RegisterLogic] = (*RegisterLogic)(nil)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	req    *types.Request
	tx     sip.ServerTransaction
}

func (l *RegisterLogic) New(ctx context.Context, svcCtx *types.ServiceContext, req *types.Request, tx sip.ServerTransaction) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		req:    req,
		tx:     tx,
	}
}

func (l *RegisterLogic) DO() *types.Response {
	if len(l.req.ID) < 18 {
		return nil
		// return &types.Response{
		// 	Error: types.NewErr("参数错误"),
		// 	Code:  types.StatusUnauthorized,
		// }
	}

	if len(l.req.Authorization) <= 0 {
		if l.svcCtx.Config.Sip.UsePassword {
			return &types.Response{
				Code:  types.StatusUnauthorized,
				Error: types.NewErr("参数错误 Authorization为空"),
			}
		}
	}

	var (
		headerExpires = l.req.Original.GetHeaders("Expires")
		expire        = 0
	)
	if len(headerExpires) > 0 {
		var (
			header = headerExpires[0]
			arr    = strings.Split(header.String(), ":")
		)
		if len(arr) == 2 {
			var (
				tmp = strings.TrimSpace(arr[1])
				err error
			)
			expire, err = strconv.Atoi(tmp)
			if err != nil {
				return &types.Response{
					Code:  types.StatusUnauthorized,
					Error: types.NewErr("expire 已过期"),
				}
			}
		}
	}

	// 检测密码
	if l.svcCtx.Config.Sip.UsePassword {
		if err := l.auth(l.req.Authorization[0], l.svcCtx.Config.Sip.Password); err != nil {
			return &types.Response{
				Code:  types.StatusUnauthorized,
				Error: types.NewErr(err.Error()),
			}
		}
	}

	var (
		now = functions.NewTimer().Now()
		// 注册地址
		record = &devices.Item{
			Devices: &devices.Devices{
				Name:           l.req.Source,
				AccessProtocol: devices.AccessProtocol_4,
				DeviceUniqueId: l.req.ID,
				State:          1,
				Online:         1,
				Expire:         uint64(now) + uint64(expire),
				Address:        l.req.Source,
				RegisterAt:     uint64(now),
			},
		}
		// 先回复 200 OK
		res = types.Response{
			BeforeResponse: func(resp sip.Response) sip.Response {
				authenticateHeader, ok := l.req.Authorization[0].(*sip.GenericHeader)
				if !ok {
					return resp
				}

				var auth = sip.AuthFromValue(authenticateHeader.Contents)
				auth.SetPassword(l.svcCtx.Config.Sip.Password).SetMethod(string(sip.REGISTER))
				if auth.CalcResponse() == auth.Response() {
					// 回应
					var (
						expires   = sip.Expires(expire)
						userAgent = sip.UserAgentHeader(l.svcCtx.Config.Name)
						server    = sip.ServerHeader(l.svcCtx.Config.Name)
					)

					resp.RemoveHeader("Allow")
					// resp.RemoveHeader("Server")
					resp.AppendHeader(&userAgent)
					resp.AppendHeader(&expires)
					resp.AppendHeader(&server)
				}

				return resp
			},
		}
	)

	if expire == 0 {
		// 设备离线
		record.Online = 0
	}

	// 注册设备
	if _, err := response.NewRpcToHttpResp[*deviceservice.Response, uint64]().Parse(
		func() (*deviceservice.Response, error) {
			item, err := record.ToMap()
			if err != nil {
				return nil, err
			}

			data, err := structpb.NewStruct(item)
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.DeviceUpsert(l.ctx, &deviceservice.MapReq{Data: data})
		},
	); err != nil {
		functions.LogError("rpc device upsert error: ", err.Error)
		return &types.Response{Error: types.NewErr(err.Error)}
	}

	//  下线后删除定时器相关记录
	if record.Online == 0 {
		// 删除catalog任务
		l.svcCtx.SipCatalogLoop <- &types.SipCatalogLoopReq{
			Req:    l.req,
			Online: false,
			Now:    now,
		}

		// 删除 心跳任务
		l.svcCtx.SipHeartbeatLoop <- &types.SipHeartbeatLoopReq{
			ID: record.DeviceUniqueId,
		}

		functions.LogInfo(fmt.Sprintf("设备下线: %s, source: %s, ID: %s", l.req.DeviceAddr.Uri, l.req.Source, l.req.ID))
		return &res
	}

	go func() {
		time.Sleep(1 * time.Second)
		// 注册定时发送catalog任务
		l.svcCtx.SipCatalogLoop <- &types.SipCatalogLoopReq{
			Req:    l.req,
			Online: true,
			Now:    now,
		}

		// 发送catalog请求
		l.req.Caller = functions.CallerFile(1)
		l.svcCtx.SipSendCatalog <- l.req
		// 发送device info请求
		l.svcCtx.SipSendDeviceInfo <- l.req

		// 设置检测心跳
		l.svcCtx.SipHeartbeatLoop <- &types.SipHeartbeatLoopReq{
			ID:               record.DeviceUniqueId,
			Now:              now,
			RegisterExpireAt: now + int64(expire),
		}
	}()

	functions.LogInfo(fmt.Sprintf("设备上线: %s, source: %s, ID: %s", l.req.DeviceAddr.Uri, l.req.Source, l.req.ID))
	return &res
}

func (l *RegisterLogic) auth(authorization sip.Header, sipPassword string) error {
	authenticateHeader, ok := authorization.(*sip.GenericHeader)
	if !ok {
		return types.NewErr("类型参数错")
	}

	var auth = sip.AuthFromValue(authenticateHeader.Contents)
	auth.SetPassword(sipPassword).SetMethod(string(sip.REGISTER))
	if auth.CalcResponse() != auth.Response() {
		return types.NewErr("密码错误")
	}

	return nil
}
