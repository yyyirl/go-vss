package gbs_sip

import (
	"net/http"
	"strings"

	"github.com/ghettovoice/gosip/sip"

	gbssip "skeyevss/core/app/sev/vss/internal/logic/gbs_sip"
	sip2 "skeyevss/core/app/sev/vss/internal/pkg/sip"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
)

func RegisterHandlers(svcCtx *types.ServiceContext) types.HType {
	return types.HType{
		sip.REGISTER: func(req sip.Request, tx sip.ServerTransaction) {
			sip2.DO("GBS", svcCtx, req, tx, nil, new(gbssip.RegisterLogic))
		},
		sip.INVITE: func(req sip.Request, tx sip.ServerTransaction) {
			sip2.DO("GBS", svcCtx, req, tx, nil, new(gbssip.InviteLogic))
		},
		sip.ACK: func(req sip.Request, tx sip.ServerTransaction) {
			sip2.DO("GBS", svcCtx, req, tx, nil, new(gbssip.ACKLogic))
		},
		sip.BYE: func(req sip.Request, tx sip.ServerTransaction) {
			sip2.DO("GBS", svcCtx, req, tx, nil, new(gbssip.ByeLogic))
		},
		sip.MESSAGE: func(req sip.Request, tx sip.ServerTransaction) {
			data, err := sip2.NewParser[types.MessageReceiveBase]().ToData(req)
			if err != nil {
				if err := tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, "MESSAGE parse failed", "")); err != nil {
					functions.LogError("Respond err:", err.Error())
				}
				return
			}

			var cmdType = data.GetCmdType()
			switch cmdType {
			case strings.ToLower(types.MessageCMDTypeKeepalive):
				sip2.DO("GBS", svcCtx, req, tx, data, new(gbssip.KeepaliveLogic))

			case strings.ToLower(types.MessageCMDTypeCatalog):
				sip2.DO("GBS", svcCtx, req, tx, data, new(gbssip.CatLogLogic))

			case strings.ToLower(types.MessageCMDTypeDeviceInfo):
				sip2.DO("GBS", svcCtx, req, tx, data, new(gbssip.DeviceInfoLogic))

			case strings.ToLower(types.MessageCMDTypeConfigDownload):
				sip2.DO("GBS", svcCtx, req, tx, data, new(gbssip.ConfigDownloadLogic))

			case strings.ToLower(types.MessageCMDTypeDeviceConfig):
				sip2.DO("GBS", svcCtx, req, tx, data, new(gbssip.DeviceConfigLogic))

			case strings.ToLower(types.MessageCMDTypePresetQuery):
				// TODO 完整版请联系作者

			case strings.ToLower(types.MessageCMDTypeRecordInfo):
				sip2.DO("GBS", svcCtx, req, tx, data, new(gbssip.VideoRecordsLogic))

			case strings.ToLower(types.MessageCMDTypeAlarm):
				sip2.DO("GBS", svcCtx, req, tx, data, new(gbssip.AlarmLogic))

			case strings.ToLower(types.MessageCMDTypeMediaStatus):
				sip2.DO("GBS", svcCtx, req, tx, data, new(gbssip.MediaStatusLogic))

			case strings.ToLower(types.MessageCMDTypeBroadcast):
				sip2.DO("GBS", svcCtx, req, tx, data, new(gbssip.BroadcastLogic))

			default:
				svcCtx.SipLog <- &types.SipLogItem{
					Content: strings.TrimSuffix(req.String(), "\n"),
					Type:    types.BroadcastTypeSipReceive,
				}
			}
		},
	}
}
