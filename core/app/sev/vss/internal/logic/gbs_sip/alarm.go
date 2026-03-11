package gbs_sip

import (
	"context"

	gosip "github.com/ghettovoice/gosip/sip"
	"google.golang.org/protobuf/types/known/structpb"

	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/app/sev/vss/internal/pkg/sip"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/alarms"
	"skeyevss/core/repositories/models/cascade"
	"skeyevss/core/repositories/models/channels"
)

var _ types.SipReceiveHandleLogic[*AlarmLogic] = (*AlarmLogic)(nil)

type AlarmLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	req    *types.Request
	tx     gosip.ServerTransaction
}

func (l *AlarmLogic) New(ctx context.Context, svcCtx *types.ServiceContext, req *types.Request, tx gosip.ServerTransaction) *AlarmLogic {
	return &AlarmLogic{
		svcCtx: svcCtx,
		ctx:    ctx,
		req:    req,
		tx:     tx,
	}
}

func (l *AlarmLogic) DO() *types.Response {
	data, err := sip.NewParser[types.SipMessageAlarm]().ToData(l.req.Original)
	if err != nil {
		return &types.Response{Error: types.NewErr(err.Error())}
	}

	from, ok := l.req.Original.From()
	if !ok {
		return &types.Response{Error: types.NewErr("from is required")}
	}

	// 检测通道信息
	channelRes, err1 := response.NewRpcToHttpResp[*deviceservice.Response, *channels.Item]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := conv.New(l.svcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: channels.ColumnUniqueId, Value: data.DeviceID},
					{Column: channels.ColumnDeviceUniqueId, Value: from.Address.User().String()},
				},
			})
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.ChannelRowFind(l.ctx, data)
		},
	)
	if err1 != nil {
		return &types.Response{Error: types.NewErr(err1.Error)}
	}

	if channelRes.Data.CascadeChannelUniqueId != "" {
		var cascadeRecords []*cascade.Item
		for _, item := range l.svcCtx.CascadeRecords {
			if item.Online <= 0 || item.State <= 0 {
				continue
			}

			var exists = false
			for _, v := range item.Relations {
				if v.Parental {
					continue
				}

				if v.UniqueId != channelRes.Data.CascadeChannelUniqueId {
					continue
				}

				exists = true
			}

			if !exists {
				continue
			}

			cascadeRecords = append(cascadeRecords, item)
		}

		// 向上级发送消息
		// TODO 完整版请联系作者
	}

	var createdAt = functions.NewTimer().NowMilli()
	if v, err := functions.NewTimer().FormatDateToMilli(data.AlarmTime, "ymdT"); err != nil {
		createdAt = v
	}

	var eventType uint = 0
	if data.Info != nil && data.Info.AlarmTypeParam != nil {
		eventType = data.Info.AlarmTypeParam.EventType
	}

	var alarmType uint = 0
	if data.Info != nil {
		switch data.AlarmMethod {
		case 2:
			if data.Info.AlarmType == 1 {
				alarmType = alarms.AlarmType_1
			} else if data.Info.AlarmType == 2 {
				alarmType = alarms.AlarmType_2
			} else if data.Info.AlarmType == 3 {
				alarmType = alarms.AlarmType_3
			} else if data.Info.AlarmType == 4 {
				alarmType = alarms.AlarmType_4
			} else if data.Info.AlarmType == 5 {
				alarmType = alarms.AlarmType_5
			}
		case 5:
			if data.Info.AlarmType == 1 {
				alarmType = alarms.AlarmType_6
			} else if data.Info.AlarmType == 2 {
				alarmType = alarms.AlarmType_7
			} else if data.Info.AlarmType == 3 {
				alarmType = alarms.AlarmType_8
			} else if data.Info.AlarmType == 4 {
				alarmType = alarms.AlarmType_9
			} else if data.Info.AlarmType == 5 {
				alarmType = alarms.AlarmType_10
			} else if data.Info.AlarmType == 6 {
				alarmType = alarms.AlarmType_11
			} else if data.Info.AlarmType == 7 {
				alarmType = alarms.AlarmType_12
			} else if data.Info.AlarmType == 8 {
				alarmType = alarms.AlarmType_13
			} else if data.Info.AlarmType == 9 {
				alarmType = alarms.AlarmType_14
			} else if data.Info.AlarmType == 10 {
				alarmType = alarms.AlarmType_15
			} else if data.Info.AlarmType == 11 {
				alarmType = alarms.AlarmType_16
			} else if data.Info.AlarmType == 12 {
				alarmType = alarms.AlarmType_17
			}
		case 6:
			if data.Info.AlarmType == 1 {
				alarmType = alarms.AlarmType_18
			} else if data.Info.AlarmType == 2 {
				alarmType = alarms.AlarmType_19
			}
		}
	}

	var record = &alarms.Item{
		Alarms: &alarms.Alarms{
			DeviceUniqueId:   data.DeviceID,
			AlarmMethod:      data.AlarmMethod,
			AlarmPriority:    data.AlarmPriority,
			AlarmDescription: data.AlarmDescription,
			Longitude:        data.Longitude,
			Latitude:         data.Latitude,
			AlarmType:        alarmType,
			EventType:        eventType,
			CreatedAt:        uint64(createdAt),
		},
	}

	// 创建报警记录
	if _, err := response.NewRpcToHttpResp[*deviceservice.Response, uint64]().Parse(
		func() (*deviceservice.Response, error) {
			data, err := structpb.NewStruct(record.ToMap())
			if err != nil {
				return nil, err
			}

			return l.svcCtx.RpcClients.Device.AlarmCreate(l.ctx, &deviceservice.MapReq{Data: data})
		},
	); err != nil {
		functions.LogError("报警记录创建失败 error: ", err.Error)
		return &types.Response{Error: types.NewErr(err.Error)}
	}

	return nil
}
