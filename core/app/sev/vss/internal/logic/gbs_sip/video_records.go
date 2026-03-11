package gbs_sip

import (
	"context"

	gosip "github.com/ghettovoice/gosip/sip"

	"skeyevss/core/app/sev/vss/internal/pkg/sip"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/repositories/models/cascade"
)

var _ types.SipReceiveHandleLogic[*VideoRecordsLogic] = (*VideoRecordsLogic)(nil)

type VideoRecordsLogic struct {
	ctx    context.Context
	svcCtx *types.ServiceContext
	req    *types.Request
	tx     gosip.ServerTransaction
}

func (l *VideoRecordsLogic) New(ctx context.Context, svcCtx *types.ServiceContext, req *types.Request, tx gosip.ServerTransaction) *VideoRecordsLogic {
	return &VideoRecordsLogic{
		svcCtx: svcCtx,
		ctx:    ctx,
		req:    req,
		tx:     tx,
	}
}

func (l *VideoRecordsLogic) DO() *types.Response {
	res, err := sip.NewParser[types.SipMessageVideoRecordsResp]().ToData(l.req.Original)
	if err != nil {
		return &types.Response{Error: types.NewErr(err.Error())}
	}

	// GBC 转发
	if v, ok := l.svcCtx.GBCRecordInfoSendMaps.Get(res.DeviceID); ok {
		defer l.svcCtx.GBCRecordInfoSendMaps.Remove(res.DeviceID)

		if v.Channel.CascadeChannelUniqueId == "" {
			return nil
		}

		var cascadeRecords []*cascade.Item
		for _, item := range l.svcCtx.CascadeRecords {
			if item.Online <= 0 || item.State <= 0 {
				continue
			}

		innerLoop:
			for _, val := range item.Relations {
				if val.Parental {
					continue
				}

				if val.UniqueId == v.Channel.CascadeChannelUniqueId {
					cascadeRecords = append(cascadeRecords, item)
					break innerLoop
				}
			}
		}

		for range cascadeRecords {
			// TODO 完整版请联系作者
			// 向上级发送消息
		}

		return nil
	}

	var records []*types.SipVideoRecordItem
	for _, item := range res.RecordList {
		var channelId = item.DeviceID
		item.DeviceID = l.req.ID
		records = append(records, &types.SipVideoRecordItem{
			SipMessageVideoRecordItem: item,
			ChannelID:                 channelId,
		})
	}

	var (
		record   = &types.SipMessageVideoRecords{Total: res.SumNum}
		cacheKey = sip.VideoRecordMapKey(l.req.ID, res.DeviceID, res.SN)
	)
	data, ok := l.svcCtx.SipMessageVideoRecordMap.Get(cacheKey)
	if ok {
		record.List = data.List
	}

	record.List = append(record.List, records...)
	l.svcCtx.SipMessageVideoRecordMap.Set(cacheKey, record)
	return nil
}
