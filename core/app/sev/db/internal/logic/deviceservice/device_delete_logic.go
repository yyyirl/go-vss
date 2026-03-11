package deviceservicelogic

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
)

type DeviceDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeviceDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeviceDeleteLogic {
	return &DeviceDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeviceDeleteLogic) DeviceDelete(in *db.XRequestParams) (*db.Response, error) {
	params, err := conv.New(l.svcCtx.Config.Mode).ToOrmParams(in)
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	list, err := l.svcCtx.DevicesModel.List(&orm.ReqParams{
		Conditions: params.Conditions,
		All:        true,
	})
	if err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	if err := response.NewMakeRpcRetErr(l.svcCtx.DevicesModel.DeleteBy(params), 2); err != nil {
		return nil, err
	}

	// 删除通道
	var uniqueIds []string
	for _, item := range list {
		uniqueIds = append(uniqueIds, item.DeviceUniqueId)
	}

	if err := l.svcCtx.ChannelsModel.DeleteBy(&orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: channels.ColumnDeviceUniqueId, Values: functions.SliceToSliceAny(uniqueIds)},
		},
	}); err != nil {
		return nil, response.NewMakeRpcRetErr(err, 2)
	}

	// var redisClient = redis.NewStreamKeepaliveRunningState(l.svcCtx.RedisClient)
	// keys, err := redisClient.Keys()
	// if err != nil {
	// 	return nil, response.NewMakeRpcRetErr(err, 2)
	// }
	//
	// var prefixDeviceUniqueIds []string
	// for _, item := range uniqueIds {
	// 	prefixDeviceUniqueIds = append(prefixDeviceUniqueIds, stream.New().PrefixDeviceUniqueId(item))
	// }

	// var deleteKeys []string
	// for _, item := range keys {
	// 	for _, v := range prefixDeviceUniqueIds {
	// 		if strings.HasPrefix(item, v) {
	// 			deleteKeys = append(deleteKeys, item)
	// 		}
	// 	}
	// }

	// return response.NewRpcResp[*db.Response]().Make(deleteKeys, 3, func(data []byte) *db.Response {
	// 	return &db.Response{Data: data, License: l.ctx.Value(interceptor.RpcReqCtxLicenseKey).(string)}
	// })
	return &db.Response{Data: []byte(strconv.FormatBool(true))}, nil
}
