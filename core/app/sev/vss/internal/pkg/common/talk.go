// @Title        talk
// @Description  main
// @Create       yiyiyi 2025/12/31 10:43

package common

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/ghettovoice/gosip/sip"

	"skeyevss/core/app/sev/vss/internal/logic/ws"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/audio"
	"skeyevss/core/pkg/sdp"
)

func SetTalkRtpConnInfo(svcCtx *types.ServiceContext, sdpInfo *sdp.Session, talkSipKey string, usablePort int) error {
	talkSipData, ok := svcCtx.TalkSipData.Get(talkSipKey)
	if !ok {
		return errors.New("sip[cache]获取失败")
	}

	if len(sdpInfo.Media) <= 0 {
		return errors.New("sdp media 获取失败 media 为空")
	}

	var (
		mediaItem = sdpInfo.Media[0]
		ssrc      = mediaItem.SSRC

		audioCodec       = audio.AudioCodecPcma
		audioPayloadType = audio.AudioPayloadTypeDefault
		audioSampleRate  = audio.AudioSampleRateDefault
	)

	for _, item := range sdpInfo.Media {
		for _, val := range item.Formats {
			if strings.ToUpper(val.Name) == audio.AudioCodecPcma || strings.ToUpper(val.Name) == audio.AudioCodecPcmu || strings.ToUpper(val.Name) == audio.AudioCodecAAC {
				audioCodec = strings.ToUpper(val.Name)
				audioPayloadType = val.Payload
				audioSampleRate = val.ClockRate
				break
			}

			if strings.ToUpper(val.Name) == audio.AudioCodecPs {
				audioCodec = strings.ToUpper(val.Name)
				audioPayloadType = val.Payload
				audioSampleRate = val.ClockRate
			}
		}
	}

	if len(mediaItem.SSRC) >= 10 {
		ssrc = mediaItem.SSRC[1:]
	}

	if ssrc == "" {
		for _, item := range sdpInfo.Media {
			mediaItem = item
			if len(mediaItem.SSRC) >= 10 {
				ssrc = mediaItem.SSRC[1:]
			}

			if ssrc != "" {
				break
			}
		}
	}

	ssrcCode, err := strconv.ParseUint(ssrc, 10, 64)
	if err != nil {
		return errors.New("ssrc 转换失败, err: " + err.Error())
	}

	talkSipData.SSRC = uint32(ssrcCode)
	talkSipData.RTPUsablePort = usablePort
	talkSipData.AudioCodec = audioCodec
	talkSipData.AudioPayloadType = audioPayloadType
	talkSipData.AudioSampleRate = audioSampleRate
	talkSipData.RTPRemoteIP = sdpInfo.Connection.Address
	talkSipData.RTPRtpPort = mediaItem.Port
	talkSipData.RTPRtcpPort = talkSipData.RTPRtpPort + 1
	svcCtx.TalkSipData.Set(talkSipKey, talkSipData)

	return nil
}

func SetTalkRtpConn(svcCtx *types.ServiceContext, req sip.Request, key string, item *audio.TalkSessionItem) error {
	// 创建RTP链接
	rtpSession, err := audio.NewRTPSession(item)
	if err != nil {
		// 停止对讲
		ws.RGBSTalkAudioStop(svcCtx, key)
		// 广播错误消息
		return errors.New("rpt会话创建失败, err: " + err.Error())
	}

	if v, ok := req.CallID(); ok {
		item.CallID = v.String()
	}
	item.Status = true
	item.ActivateAt = time.Now().UnixMilli()
	item.RTPSession = rtpSession
	item.ACKReq = req
	svcCtx.TalkSipData.Set(key, item)

	// 广播客户端sip状态
	ws.BGBSSendTalkPub(svcCtx, key, 3)
	// 广播占用状态已被解除 初始化
	ws.BGBSSendTalkUsageStatus(svcCtx, "", key, 0)
	return nil
}
