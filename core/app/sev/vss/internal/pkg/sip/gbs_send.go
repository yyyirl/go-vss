/**
 * @Author:         yi
 * @Description:    send
 * @Version:        1.0.0
 * @Date:           2025/6/20 15:46
 */
package sip

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ghettovoice/gosip/sip"

	"skeyevss/core/app/sev/vss/internal/pkg/rule"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/common/stream"
	"skeyevss/core/pkg/audio"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/sdp"
	"skeyevss/core/repositories/models/devices"
)

var ssrcCounter int64 = 200000001

type GBSSender struct {
	svcCtx *types.ServiceContext
	req    *types.Request

	deviceUniqueId string // 区分设备id/通道id

	setting *rule.Config
}

func NewGBSSender(svcCtx *types.ServiceContext, req *types.Request, DeviceUniqueId string) *GBSSender {
	return &GBSSender{
		svcCtx:         svcCtx,
		req:            req,
		setting:        rule.NewConfig(svcCtx.Config, svcCtx.Setting).Conv(),
		deviceUniqueId: DeviceUniqueId,
	}
}

func (l *GBSSender) makeHeader(Type headerType) sip.Header {
	return l.doMakeHeader(Type, nil)
}

func (l *GBSSender) makeHeaderWithBody(Type headerType, body string) sip.Header {
	return l.doMakeHeader(Type, body)
}

func (l *GBSSender) makeHeaderWith(Type headerType, body interface{}) sip.Header {
	return l.doMakeHeader(Type, body)
}

func (l *GBSSender) doMakeHeader(Type headerType, data interface{}) sip.Header {
	switch Type {
	case headerTypeVia:
		var port = uint16(l.svcCtx.Config.Sip.Port)
		return sip.ViaHeader{
			{
				ProtocolName:    "SIP",
				ProtocolVersion: "2.0",
				Transport:       l.req.TransportProtocol,
				Host:            l.setting.SipIP(),
				Port:            (*sip.Port)(&port),
				Params: sip.NewParams().
					Add("rport", sip.String{Str: strconv.Itoa(int(port))}).
					Add("branch", sip.String{Str: "z9hG4bK" + functions.RandWithString("0123456789", 10)}),
			},
		}

	case headerTypeFrom:
		var (
			port     = uint16(l.svcCtx.Config.Sip.Port)
			fromAddr = sip.Address{
				DisplayName: l.req.DeviceAddr.DisplayName,
				Uri: &sip.SipUri{
					FUser: sip.String{Str: l.svcCtx.Config.Sip.ID},
					FHost: l.setting.SipIP(),
					FPort: (*sip.Port)(&port),
				},
				Params: sip.NewParams().Add("tag", sip.String{Str: functions.RandWithString("0123456789", 9)}),
			}
		)
		return fromAddr.AsFromHeader()

	case headerTypeTo:
		return l.toAddress().AsToHeader()

	case headerTypeToWith:
		var toHeader = l.toAddress().AsToHeader()
		toHeader.Address.SetUser(sip.String{Str: data.(string)})
		return toHeader

	case headerTypeCallId:
		var callId = sip.CallID(functions.RandWithString("0123456789", 10))
		return &callId

	case headerTypeCallIdWith:
		var callId = data.(sip.CallID)
		return &callId

	case headerTypeUserAgent:
		var userAgent = sip.UserAgentHeader(MakeCascadeUserAgent(l.svcCtx.Config.Name, l.svcCtx.Config.InternalIp))
		return &userAgent

	case headerTypeMessageCSEq:
		var csEq = sip.CSeq{
			SeqNo:      l.SN(l.deviceUniqueId),
			MethodName: sip.MESSAGE,
		}
		if data != nil {
			if v, ok := data.(sip.RequestMethod); ok {
				csEq = sip.CSeq{
					SeqNo:      l.SN(l.deviceUniqueId),
					MethodName: v,
				}
			}
		}

		return &csEq

	case headerTypeMaxForwards:
		var maxForwards = sip.MaxForwards(70)
		return &maxForwards

	case headerTypeContentType:
		return &sip.GenericHeader{
			HeaderName: HeaderContentType,
			Contents:   "Application/MANSCDP+xml",
		}

	case headerTypeContentTypeSDP:
		return &sip.GenericHeader{
			HeaderName: HeaderContentType,
			Contents:   "APPLICATION/SDP",
		}

	case headerTypeContentTypeMANSRTSP:
		return &sip.GenericHeader{
			HeaderName: HeaderContentType,
			Contents:   "Application/MANSRTSP",
		}

	case headerTypeContentLength:
		return &sip.GenericHeader{
			HeaderName: HeaderContentLength,
			Contents:   strconv.Itoa(len(data.(string))),
		}

	case headerTypeExpire:
		return &sip.GenericHeader{
			HeaderName: "Expires", Contents: "3900",
		}

	case headerTypeContact:
		var contact = &sip.ContactHeader{DisplayName: l.req.DeviceAddr.DisplayName}
		if l.req.DeviceAddr.Uri != nil {
			contact.Address = l.req.DeviceAddr.Uri.Clone()
		}

		return contact

	case headerTypeContactCurrent:
		var (
			port    = uint16(l.svcCtx.Config.Sip.Port)
			contact = &sip.ContactHeader{
				Address: &sip.SipUri{
					FUser: sip.String{Str: l.svcCtx.Config.Sip.ID},
					FHost: l.setting.SipIP(),
					FPort: (*sip.Port)(&port),
				},
			}
		)

		return contact

	case headerTypeEventPresence:
		return &sip.GenericHeader{HeaderName: "Event", Contents: "presence"}

	case headerTypeEventCatalog:
		return &sip.GenericHeader{
			HeaderName: "Event", Contents: fmt.Sprintf("Catalog;id=%s", functions.RandWithString("0123456789", 9)),
		}

	case headerTypeSubject:
		if data == nil {
			return nil
		}

		v, ok := data.(string)
		if !ok {
			return nil
		}

		return &sip.GenericHeader{HeaderName: "Subject", Contents: v}

	default:
		return nil
	}
}

func (l *GBSSender) toAddress() *sip.Address {
	// 非同一域的目标地址需要使用@host
	if l.deviceUniqueId[0:9] != l.svcCtx.Config.Sip.Domain {
		return &sip.Address{
			Uri: &sip.SipUri{
				FUser: sip.String{Str: l.deviceUniqueId},
				FHost: l.req.Source,
			},
		}
	}

	return &sip.Address{
		Uri: &sip.SipUri{
			FUser: sip.String{Str: l.deviceUniqueId},
			FHost: l.svcCtx.Config.Sip.Domain,
		},
	}
}

func (l *GBSSender) makeRequestBody(data interface{}) (string, error) {
	body, err := xml.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		`<?xml version="1.0" encoding="GB2312"?>
%s`,
		body,
	), nil
}

func (l *GBSSender) makeRequest(method sip.RequestMethod, headers []sip.Header, body string) sip.Request {
	var (
		uri = l.toAddress().Uri
		req = sip.NewRequest(
			"",
			method,
			uri,
			"SIP/2.0",
			functions.ArrFilter(headers, func(item sip.Header) bool {
				return item != nil
			}),
			body,
			nil,
		)
	)
	req.SetTransport(l.req.TransportProtocol)
	if v := uri.Port(); v != nil {
		req.SetDestination(fmt.Sprintf("%s:%d", uri.Host(), v.String()))
	} else {
		req.SetDestination(uri.Host())
	}

	return req
}

func (l *GBSSender) SN(uniqueId string) uint32 {
	sn, ok := l.svcCtx.SipGBSSNMap.Get(uniqueId)
	if !ok {
		sn = 1
	} else {
		sn = sn + 1
	}

	l.svcCtx.SipGBSSNMap.Set(uniqueId, sn)
	return sn
}

func (l *GBSSender) Send(data sip.Request) (sip.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(l.svcCtx.Config.Sip.SendTimeout)*time.Second)
	defer cancel()

	var content = data.String()
	l.svcCtx.SipLog <- &types.SipLogItem{
		Content: content,
		Type:    types.BroadcastTypeSipRequest,
	}

	if l.svcCtx.Config.UseSipPrintLog {
		SipLog(l.svcCtx.Config.UseSipLogToFile, l.svcCtx.Config.SipLogPath, types.BroadcastTypeSipRequest, "GBS", data, content)
	}

	var (
		resp sip.Response
		err  error
	)
	if l.req.TransportProtocol == "UDP" {
		resp, err = (*l.svcCtx.GBSUDPSev).RequestWithContext(ctx, data)
	} else {
		resp, err = (*l.svcCtx.GBSTCPSev).RequestWithContext(ctx, data)
	}

	if l.svcCtx.Config.UseSipPrintLog {
		if err != nil {
			requestError, ok := err.(*sip.RequestError)
			if ok && requestError.Response != nil {
				SipLog(l.svcCtx.Config.UseSipLogToFile, l.svcCtx.Config.SipLogPath, types.BroadcastTypeSipReceiveResponse, "GBS", data, requestError.Response.String())
			}
		} else {
			SipLog(l.svcCtx.Config.UseSipLogToFile, l.svcCtx.Config.SipLogPath, types.BroadcastTypeSipReceiveResponse, "GBS", data, resp.String())
		}
	}

	return resp, err
}

func (l *GBSSender) SendDirect(data sip.Request) error {
	var content = data.String()
	l.svcCtx.SipLog <- &types.SipLogItem{
		Content: content,
		Type:    types.BroadcastTypeSipRequest,
	}

	if l.svcCtx.Config.UseSipPrintLog {
		SipLog(l.svcCtx.Config.UseSipLogToFile, l.svcCtx.Config.SipLogPath, types.BroadcastTypeSipRequest, "GBS", data, content)
	}

	if l.req.TransportProtocol == "UDP" {
		return (*l.svcCtx.GBSUDPSev).Send(data)
	}

	return (*l.svcCtx.GBSTCPSev).Send(data)
}

// 主动响应
func (l *GBSSender) response(tx sip.ServerTransaction, req sip.Request, resp sip.Response) error {
	var content = resp.String()
	l.svcCtx.SipLog <- &types.SipLogItem{
		Content: content,
		Type:    types.BroadcastTypeSipResponse,
	}

	if tx == nil {
		return errors.New("tx is nil")
	}

	if l.svcCtx.Config.UseSipPrintLog {
		SipLog(l.svcCtx.Config.UseSipLogToFile, l.svcCtx.Config.SipLogPath, types.BroadcastTypeSipResponse, "GBS", req, content)
	}

	return tx.Respond(resp)
}

func (l *GBSSender) MakeSDPResponse(req sip.Request, statusCode sip.StatusCode, reason string, body string) sip.Response {
	var resp = sip.NewResponse(
		req.MessageID(),
		req.SipVersion(),
		statusCode,
		reason,
		[]sip.Header{},
		body,
		req.Fields(),
	)
	sip.CopyHeaders("Record-Route", req, resp)
	sip.CopyHeaders("Via", req, resp)
	sip.CopyHeaders("From", req, resp)

	if v, ok := req.To(); ok {
		v.Params.Add("tag", sip.String{Str: functions.RandWithString("0123456789", 9)})
		resp.AppendHeader(v)
	} else {
		sip.CopyHeaders("To", req, resp)
	}

	sip.CopyHeaders("Call-ID", req, resp)
	sip.CopyHeaders("CSeq", req, resp)
	resp.AppendHeader(l.makeHeader(headerTypeContactCurrent))

	if statusCode == 100 {
		sip.CopyHeaders("Timestamp", req, resp)
	}

	var contentType = sip.ContentType("Application/SDP")
	resp.AppendHeader(&contentType)

	resp.SetBody(body, true)
	resp.SetTransport(req.Transport())
	resp.SetSource(req.Destination())
	resp.SetDestination(req.Source())

	return resp
}

// Invite ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) VideoLiveInvite(data *types.SipVideoLiveInviteMessage) (sip.Request, sip.Response, error) {
	var (
		headers = []sip.Header{
			l.makeHeader(headerTypeVia),
			l.makeHeader(headerTypeFrom),
			l.makeHeader(headerTypeTo),
			l.makeHeader(headerTypeCallId),
			l.makeHeader(headerTypeUserAgent),
			l.makeHeaderWith(headerTypeMessageCSEq, sip.INVITE),
			l.makeHeader(headerTypeMaxForwards),
			l.makeHeader(headerTypeContentTypeSDP),
			l.makeHeader(headerTypeContactCurrent),
		}
		ssrc       = strings.TrimSpace(data.ChannelUniqueId[3:8] + data.ChannelUniqueId[16:20])
		playFlag   = 0
		isPlayback = data.StartAt != "" && data.EndAt != "" && data.PlayType == stream.PlayTypePlayback
	)
	if data.PlayType == stream.PlayTypePlayback {
		playFlag = 1
		atomic.AddInt64(&ssrcCounter, 1)
		ssrc = strconv.FormatInt(ssrcCounter, 10)
	}

	if isPlayback {
		headers = append(headers, l.makeHeaderWithBody(headerTypeSubject, fmt.Sprintf("%s:%d%s,%s:%s", data.ChannelUniqueId, playFlag, ssrc, l.svcCtx.Config.Sip.ID, ssrc)))
	} else {
		headers = append(headers, l.makeHeaderWithBody(headerTypeSubject, fmt.Sprintf("%s:%d%s,%s:0", data.ChannelUniqueId, playFlag, ssrc, l.svcCtx.Config.Sip.ID)))
	}

	var proto = "TCP/RTP/AVP"
	if data.TransportProtocol.MediaProtocolMode == 0 {
		proto = "RTP/AVP"
	}

	var sdpInfo = &sdp.Session{
		Version: 0,
		Origin: &sdp.Origin{
			Username:       data.ChannelUniqueId,
			Address:        data.MediaServerIP,
			SessionID:      0,
			SessionVersion: 0,
		},
		Name:       functions.Capitalize(string(data.PlayType)),
		Connection: &sdp.Connection{Address: data.MediaServerIP},
		Media: []*sdp.Media{
			{
				Type:  "video",
				Port:  int(data.StreamPort),
				Proto: proto,
				Mode:  sdp.ModeRecvOnly,
				Formats: []*sdp.Format{
					{Payload: 96, Name: "PS", ClockRate: 90000},
					{Payload: 97, Name: "MPEG4", ClockRate: 90000},
					{Payload: 98, Name: "H264", ClockRate: 90000},
					{Payload: 99, Name: "H265", ClockRate: 90000},
				},
				SSRC: fmt.Sprintf("%d%s", playFlag, ssrc),
			},
		},
		URI: fmt.Sprintf("%s:0", data.ChannelUniqueId),
	}
	if data.TransportProtocol.MediaProtocolMode == 1 {
		sdpInfo.Media[0].Attributes = sdp.Attributes{
			sdp.NewAttr("setup", data.TransportProtocol.MediaTransMode),
			sdp.NewAttr("connection", "new"),
		}
	}

	if isPlayback {
		startAt, err := time.ParseInLocation("2006-01-02 15:04:05", data.StartAt, time.Local)
		if err != nil {
			return nil, nil, err
		}

		endAt, err := time.ParseInLocation("2006-01-02 15:04:05", data.EndAt, time.Local)
		if err != nil {
			return nil, nil, err
		}

		sdpInfo.Name = "Playback"
		sdpInfo.Timing = &sdp.Timing{
			Start: startAt,
			Stop:  endAt,
		}

		if data.Download {
			sdpInfo.Name = "Download"
			if len(sdpInfo.Media) > 0 && len(sdpInfo.Media[0].Attributes) > 0 {
				sdpInfo.Media[0].Attributes = append(
					sdpInfo.Media[0].Attributes,
					// sdp.NewAttr("downloadspeed", fmt.Sprintf("%d", data.Speed)),
					sdp.NewAttr("downloadspeed", "4"),
				)
			}
		}
	}

	if data.TransportProtocol.BitstreamIndex > 0 {
		if v, ok := devices.VBitstreamIndexes[data.TransportProtocol.BitstreamIndex]; ok {
			sdpInfo.Media[0].Attributes = append(
				sdpInfo.Media[0].Attributes,
				sdp.NewAttr(v.Key, v.Value),
			)
		}
	}

	var body = sdpInfo.String()
	headers = append(headers, l.makeHeaderWithBody(headerTypeContentLength, body))
	var request = l.makeRequest(sip.INVITE, headers, body)
	response, err := l.Send(request)
	if err != nil {
		return nil, nil, err
	}

	return request, response, nil
}

func (l *GBSSender) TalkInvite(data *types.SipTalkInviteMessage, usablePort uint) (sip.Request, sip.Response, error) {
	// TODO 完整版请联系作者
	return nil, nil, nil
}

func (l *GBSSender) InviteSDPResponse(tx sip.ServerTransaction, sdpInfo *sdp.Session, usablePort uint) error {
	sdpInfoResp, err := l.makeInviteSDPResp(l.req, sdpInfo, usablePort)
	if err != nil {
		return err
	}

	return l.response(
		tx,
		l.req.Original,
		l.MakeSDPResponse(l.req.Original, sip.StatusCode(200), "OK", sdpInfoResp.String()),
	)
}

func (l *GBSSender) makeInviteSDPResp(req *types.Request, data *sdp.Session, usablePort uint) (*sdp.Session, error) {
	var (
		mediaType       = ""
		mediaProto      = ""
		description     = SDPMediaDescription_1
		mediaAttributes = sdp.Attributes{}
		mediaFormats    = []*sdp.Format{}
		ssrc            = strings.TrimSpace(req.ID[3:8] + req.ID[16:20])

		audioCodec       = audio.AudioCodecPcma
		audioPayloadType = audio.AudioPayloadTypeDefault
		audioSampleRate  = audio.AudioSampleRateDefault
	)
	for _, item := range data.Media {
		for _, val := range item.Formats {
			if strings.ToUpper(val.Name) == audio.AudioCodecPcma || strings.ToUpper(val.Name) == audio.AudioCodecPcmu || strings.ToUpper(val.Name) == audio.AudioCodecAAC {
				audioCodec = strings.ToUpper(val.Name)
				audioPayloadType = val.Payload
				// audioSampleRate = val.ClockRate
				break
			}

			if strings.ToUpper(val.Name) == audio.AudioCodecPs {
				audioCodec = strings.ToUpper(val.Name)
				audioPayloadType = val.Payload
				// audioSampleRate = val.ClockRate
			}
		}
		if item.Type == "audio" {
			mediaType = "audio"
			mediaProto = item.Proto
			mediaAttributes = sdp.Attributes{
				sdp.NewAttr("rtpmap", fmt.Sprintf("%d %s/%d/1", audioPayloadType, strings.ToUpper(audioCodec), audioSampleRate)),
				sdp.NewAttr("sendonly", ""),
			}
			mediaFormats = []*sdp.Format{{Payload: audioPayloadType}}
			ssrc = item.SSRC
			description = item.Description
		}
	}

	if mediaType == "" || mediaProto == "" {
		return nil, errors.New("媒体信息获取失败")
	}

	var ip = l.svcCtx.Config.ExternalIp
	return &sdp.Session{
		Version: 0,
		Origin: &sdp.Origin{
			Username:       l.svcCtx.Config.Sip.ID,
			Address:        ip,
			SessionID:      0,
			SessionVersion: 0,
			Network:        "IN",
			Type:           "IP4",
		},
		Name: functions.Capitalize(string(stream.PlayTypePlay)),
		// 发送音频内网ip 当前服务器公网ip
		Connection: &sdp.Connection{Address: ip},
		Timing:     &sdp.Timing{},
		Media: []*sdp.Media{
			{
				Type:  mediaType,
				Port:  int(usablePort),
				Proto: mediaProto,
				// Mode:       sdp.ModeRecvOnly,
				Attributes:  mediaAttributes,
				SSRC:        ssrc,
				Formats:     mediaFormats,
				Description: description,
			},
		},
	}, nil
}

// Ack ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) AckReq(resp sip.Response) (sip.Request, error) {
	var callId sip.CallID
	if v, ok := resp.CallID(); ok {
		callId = *v
	}

	to, _ := resp.To()
	from, _ := resp.From()

	var request = l.makeRequest(
		sip.ACK,
		[]sip.Header{
			l.makeHeader(headerTypeVia),
			from,
			to,
			l.makeHeaderWith(headerTypeCallIdWith, callId),
			l.makeHeader(headerTypeUserAgent),
			l.makeHeaderWith(headerTypeMessageCSEq, sip.ACK),
			l.makeHeader(headerTypeMaxForwards),
			l.makeHeader(headerTypeContactCurrent),
			l.makeHeaderWithBody(headerTypeContentLength, ""),
		},
		"",
	)

	return request, nil
}

// Bye ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) Bye(ackReq sip.Request) (sip.Response, error) {
	var callId sip.CallID
	if v, ok := ackReq.CallID(); ok {
		callId = *v
	}

	to, _ := ackReq.To()
	from, _ := ackReq.From()
	return l.Send(
		l.makeRequest(
			sip.BYE,
			[]sip.Header{
				l.makeHeader(headerTypeVia),
				from,
				to,
				l.makeHeaderWith(headerTypeCallIdWith, callId),
				l.makeHeader(headerTypeUserAgent),
				l.makeHeaderWith(headerTypeMessageCSEq, sip.BYE),
				l.makeHeader(headerTypeMaxForwards),
				l.makeHeader(headerTypeContactCurrent),
				l.makeHeaderWithBody(headerTypeContentLength, ""),
			},
			"",
		),
	)
}

func (l *GBSSender) TalkBye(req sip.Request) (sip.Response, error) {
	var callId sip.CallID
	if v, ok := req.CallID(); ok {
		callId = *v
	}

	to, _ := req.To()
	from, _ := req.From()
	return l.Send(
		l.makeRequest(
			sip.BYE,
			[]sip.Header{
				l.makeHeader(headerTypeVia),
				&sip.FromHeader{
					Address: to.Address.Clone(),
					Params:  to.Params,
				},
				&sip.ToHeader{
					Address: from.Address.Clone(),
					Params:  from.Params,
				},
				l.makeHeaderWith(headerTypeCallIdWith, callId),
				l.makeHeader(headerTypeUserAgent),
				l.makeHeaderWith(headerTypeMessageCSEq, sip.BYE),
				l.makeHeader(headerTypeMaxForwards),
				l.makeHeader(headerTypeContactCurrent),
				l.makeHeaderWithBody(headerTypeContentLength, ""),
			},
			"",
		),
	)
}

// 获取设备信息 -----------------------------------------------------------------------------------------------------------------------

// rtp ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) SendWithRtpData(req *types.GBSSipSendTalk, talkSipData *audio.TalkSessionItem) error {
	if talkSipData == nil {
		return errors.New("not found SIP data")
	}

	if len(req.Data) <= 0 {
		return errors.New("invalid stream")
	}

	if talkSipData.RTPSession == nil {
		return errors.New("RTPSession 为空")
	}

	// 更新活跃时间
	talkSipData.ActivateAt = time.Now().UnixMilli()
	l.svcCtx.TalkSipData.Set(req.DeviceUniqueId, talkSipData)

	return talkSipData.RTPSession.SendAudioStream(req.Data)
}

// rtp ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) DeviceInfo() (sip.Response, error) {
	body, err := l.makeRequestBody(
		types.SipMessageGBSDeviceInfo{
			CmdType:  types.MessageCMDTypeDeviceInfo,
			DeviceID: l.deviceUniqueId,
			SN:       l.SN(l.deviceUniqueId),
		},
	)
	if err != nil {
		return nil, err
	}

	return l.Send(
		l.makeRequest(
			sip.MESSAGE,
			[]sip.Header{
				l.makeHeader(headerTypeVia),
				l.makeHeader(headerTypeFrom),
				l.makeHeader(headerTypeTo),
				l.makeHeader(headerTypeCallId),
				l.makeHeader(headerTypeUserAgent),
				l.makeHeader(headerTypeMessageCSEq),
				l.makeHeader(headerTypeMaxForwards),
				l.makeHeader(headerTypeContentType),
				l.makeHeaderWithBody(headerTypeContentLength, body),
			},
			body,
		),
	)
}

// catalog ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) Catalog() (sip.Response, error) {
	body, err := l.makeRequestBody(
		types.SipMessageGBSCatalog{
			CmdType:  types.MessageCMDTypeCatalog,
			DeviceID: l.deviceUniqueId,
			SN:       l.SN(l.deviceUniqueId),
		},
	)
	if err != nil {
		return nil, err
	}

	return l.Send(
		l.makeRequest(
			sip.MESSAGE,
			[]sip.Header{
				l.makeHeader(headerTypeVia),
				l.makeHeader(headerTypeFrom),
				l.makeHeader(headerTypeTo),
				l.makeHeader(headerTypeCallId),
				l.makeHeader(headerTypeUserAgent),
				l.makeHeader(headerTypeMessageCSEq),
				l.makeHeader(headerTypeMaxForwards),
				l.makeHeader(headerTypeContentType),
				l.makeHeaderWithBody(headerTypeContentLength, body),
			},
			body,
		),
	)
}

// 设备控制 ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) DeviceControl(ptz string) (sip.Response, error) {
	// TODO 完整版请联系作者
	return nil, nil
}

// 回放控制 ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) PlaybackControl(ackReq sip.Request, streamRes *stream.Item, req types.VideoPlaybackControlReq) (sip.Response, error) {
	callID, ok := ackReq.CallID()
	if !ok {
		return nil, errors.New("cannot send playback control")
	}

	var (
		cmdType = "SCALE"
		cmdBuf  = bytes.NewBufferString("")
	)
	switch cmdType {
	case "SCALE":
		cmdBuf.WriteString("PLAY RTSP/1.0\r\n")
		cmdBuf.WriteString(fmt.Sprintf("CSeq: %d\r\n", l.SN(streamRes.Channel)))
		cmdBuf.WriteString(fmt.Sprintf("Scale: %.2f\r\n", req.Speed))

	default:
		// err = fmt.Errorf("unknown Command[%s]", cmd)
		// return
		return nil, fmt.Errorf("unsupported command: %s", cmdType)
	}

	var body = cmdBuf.String()
	to, _ := ackReq.To()
	from, _ := ackReq.From()
	return l.Send(
		l.makeRequest(
			sip.INFO,
			[]sip.Header{
				l.makeHeader(headerTypeVia),
				from,
				to,
				l.makeHeaderWith(headerTypeCallIdWith, callID),
				l.makeHeader(headerTypeUserAgent),
				l.makeHeader(headerTypeMessageCSEq),
				l.makeHeader(headerTypeMaxForwards),
				l.makeHeader(headerTypeContentTypeMANSRTSP),
				l.makeHeaderWithBody(headerTypeContentLength, body),
			},
			body,
		),
	)
}

// 获取预置点 ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) QueryPresetPoints() (sip.Response, error) {
	// TODO 完整版请联系作者
	return nil, nil
}

// 设置预置点 ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) SetPresetPoints(ptz string) (sip.Response, error) {
	// TODO 完整版请联系作者
	return nil, nil
}

// 获取录像 ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) QueryVideoRecords(day, sn int64) (sip.Response, error) {
	var (
		date  = functions.NewTimer().FormatTimestamp(day, functions.TimeFormatYmd)
		start = fmt.Sprintf("%sT00:00:00", date)
		end   = fmt.Sprintf("%sT23:59:59", date)
	)

	body, err := l.makeRequestBody(
		types.SipMessageGBSRecordInfo{
			CmdType:   types.MessageCMDTypeRecordInfo,
			DeviceID:  l.deviceUniqueId,
			SN:        uint32(sn),
			StartTime: start,
			EndTime:   end,
			Type:      "all",
		},
	)
	if err != nil {
		return nil, err
	}

	return l.Send(
		l.makeRequest(
			sip.MESSAGE,
			[]sip.Header{
				l.makeHeader(headerTypeVia),
				l.makeHeader(headerTypeFrom),
				l.makeHeader(headerTypeTo),
				l.makeHeader(headerTypeCallId),
				l.makeHeader(headerTypeUserAgent),
				l.makeHeader(headerTypeMessageCSEq),
				l.makeHeader(headerTypeMaxForwards),
				l.makeHeader(headerTypeContentType),
				l.makeHeaderWithBody(headerTypeContentLength, body),
			},
			body,
		),
	)
}

// 布防 ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) SetGuard() (sip.Response, error) {
	body, err := l.makeRequestBody(
		types.SipMessageGBSGuard{
			CmdType:  types.MessageCMDTypeDeviceControl,
			DeviceID: l.deviceUniqueId,
			SN:       l.SN(l.deviceUniqueId),
			GuardCmd: types.MessageGuardCmdTypeSetGuard,
		},
	)
	if err != nil {
		return nil, err
	}

	return l.Send(
		l.makeRequest(
			sip.MESSAGE,
			[]sip.Header{
				l.makeHeader(headerTypeVia),
				l.makeHeader(headerTypeFrom),
				l.makeHeader(headerTypeTo),
				l.makeHeader(headerTypeCallId),
				l.makeHeader(headerTypeUserAgent),
				l.makeHeader(headerTypeMessageCSEq),
				l.makeHeader(headerTypeMaxForwards),
				l.makeHeader(headerTypeContentType),
				l.makeHeaderWithBody(headerTypeContentLength, body),
			},
			body,
		),
	)
}

// 订阅 ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) Subscription(cmdType string) (sip.Response, error) {
	var reqContent interface{}
	if cmdType == types.SubscriptionCatalog {
		reqContent = &types.SipMessageGBSSubscriptionCatalog{
			CmdType:  cmdType,
			DeviceID: l.deviceUniqueId,
			SN:       l.SN(l.deviceUniqueId),
		}
	} else if cmdType == types.SubscriptionAlarm {
		reqContent = &types.SipMessageGBSSubscriptionAlarm{
			CmdType:            cmdType,
			DeviceID:           l.deviceUniqueId,
			SN:                 l.SN(l.deviceUniqueId),
			StartAlarmPriority: 0,
			EndAlarmPriority:   0,
			AlarmMethod:        0,
		}

	} else if cmdType == types.SubscriptionMobilePosition {
		reqContent = &types.SipMessageGBSSubscriptionLocation{
			CmdType:  cmdType,
			DeviceID: l.deviceUniqueId,
			SN:       l.SN(l.deviceUniqueId),
			Interval: 5,
		}
	} else if cmdType == types.SubscriptionPTZPosition {
		reqContent = &types.SipMessageGBSPtz{
			CmdType:  cmdType,
			DeviceID: l.deviceUniqueId,
			SN:       l.SN(l.deviceUniqueId),
		}
	} else {
		return nil, errors.New("invalid cmdType")
	}

	body, err := l.makeRequestBody(reqContent)
	if err != nil {
		return nil, err
	}

	var headers = []sip.Header{
		l.makeHeader(headerTypeVia),
		l.makeHeader(headerTypeFrom),
		l.makeHeader(headerTypeTo),
		l.makeHeader(headerTypeCallId),
		l.makeHeader(headerTypeUserAgent),
		l.makeHeader(headerTypeMessageCSEq),
		l.makeHeader(headerTypeMaxForwards),
		l.makeHeader(headerTypeExpire),
		l.makeHeader(headerTypeContentType),
		l.makeHeaderWithBody(headerTypeContentLength, body),
	}
	if cmdType == types.SubscriptionCatalog {
		if l.deviceUniqueId[:10] != l.svcCtx.Config.Sip.ID[:10] {
			headers = append(
				headers,
				l.makeHeader(headerTypeContact),
				l.makeHeader(headerTypeEventCatalog),
			)
		} else {
			headers = append(headers, l.makeHeader(headerTypeEventPresence))
		}
	} else {
		if l.deviceUniqueId[:10] != l.svcCtx.Config.Sip.ID[:10] {
			headers = append(headers, l.makeHeader(headerTypeContact))
		}

		headers = append(headers, l.makeHeader(headerTypeEventPresence))
	}

	return l.Send(l.makeRequest(sip.SUBSCRIBE, headers, body))
}

// Broadcast ------------------------------------------------------------------------------------------------------------------------

func (l *GBSSender) Broadcast() (sip.Response, error) {
	// TODO 完整版请联系作者

	return nil, nil
}
