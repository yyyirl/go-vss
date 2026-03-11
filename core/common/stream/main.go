// @Title        main
// @Description  main
// @Create       yiyiyi 2025/7/14 13:54

package stream

import (
	"fmt"
	"path"
	"strings"
	"sync/atomic"

	cTypes "skeyevss/core/common/types"
	"skeyevss/core/constants"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/tps"
)

type (
	Stream struct{}

	Item struct {
		Channel,
		Device string
		PlayType PlayType
	}
)

type PlayType string

var PlaybackCount int64 = 0

const (
	PlayTypeTalk     PlayType = "talk"
	PlayTypePlay     PlayType = "play"
	PlayTypePlayback PlayType = "playback"
)

func New() *Stream {
	return &Stream{}
}

const streamNameFlag = "_"

func (s *Stream) Produce(deviceUniqueId, channelUniqueId string, playType PlayType) string {
	atomic.AddInt64(&PlaybackCount, 1)

	if playType == PlayTypePlayback {
		return fmt.Sprintf(
			"stream%s%s%s%s%s%s%s%d",
			streamNameFlag,
			deviceUniqueId,
			streamNameFlag,
			channelUniqueId,
			streamNameFlag,
			playType,
			streamNameFlag,
			PlaybackCount,
		)
	}

	// playback play/playback
	return fmt.Sprintf(
		"stream%s%s%s%s%s%s",
		streamNameFlag,
		deviceUniqueId,
		streamNameFlag,
		channelUniqueId,
		streamNameFlag,
		playType,
	)
}

func (s *Stream) ProduceWith(deviceUniqueId, channelUniqueId string, playType PlayType, uniqueId string) string {
	return fmt.Sprintf(
		"stream%s%s%s%s%s%s%s%s",
		streamNameFlag,
		deviceUniqueId,
		streamNameFlag,
		channelUniqueId,
		streamNameFlag,
		playType,
		streamNameFlag,
		uniqueId,
	)
}

func (s *Stream) Snapshot(saveVideoSnapshotDir, deviceUniqueId, channelUniqueId string) string {
	return path.Join(saveVideoSnapshotDir, fmt.Sprintf("%s-%s.jpg", deviceUniqueId, channelUniqueId))
}

func (s *Stream) PrefixDeviceUniqueId(deviceUniqueId string) string {
	return fmt.Sprintf(
		"stream%s%s",
		streamNameFlag,
		deviceUniqueId,
	)
}

func (s *Stream) Parse(streamName string) (*Item, error) {
	var arr = strings.Split(streamName, streamNameFlag)
	if len(arr) < 4 {
		return nil, fmt.Errorf("invalid stream name: %s", streamName)
	}

	return &Item{
		Device:   arr[1],
		Channel:  arr[2],
		PlayType: s.ToPlayType(arr[3]),
	}, nil
}

func (s *Stream) PlayTypeVerify(v PlayType) bool {
	return functions.Contains(string(v), []string{"play", "playback"})
}

func (s *Stream) ToPlayType(v string) PlayType {
	if PlayType(strings.ToLower(v)) == PlayTypePlayback {
		return PlayTypePlayback
	}

	return PlayTypePlay
}

func (s *Stream) PlayAddress(streamPlayProxyPath tps.YamlStreamPlayProxyPath, msNode *cTypes.MSVoteNodeResp, streamName string) *cTypes.PlayAddress {
	var (
		port           = msNode.HttpPort
		httpProtocol   = "http"
		wsProtocol     = "ws"
		webrtcProtocol = "webrtc"
	)
	if msNode.UseHttpsPlay {
		port = msNode.HttpsPort
		httpProtocol = "https"
		wsProtocol = "wss"
		webrtcProtocol = "webrtcs"
	}

	var data = &cTypes.PlayAddress{
		Rtsp: &cTypes.PlayAddressItem{
			Name: constants.VideoPlayAddressTypeRtsp,
			Url:  fmt.Sprintf("rtsp://%s:%d/live/%s", msNode.IP, msNode.RtspPort, streamName),
		},
		Rtmp: &cTypes.PlayAddressItem{
			Name: constants.VideoPlayAddressTypeRtmp,
			Url:  fmt.Sprintf("rtmp://%s:%d/live/%s", msNode.IP, msNode.RtmpPort, streamName),
		},
		HttpFlv: &cTypes.PlayAddressItem{
			Name: constants.VideoPlayAddressTypeHttpFlv,
			Url:  fmt.Sprintf("%s://%s:%d/live/%s.flv", httpProtocol, msNode.IP, port, streamName),
		},
		WSFlv: &cTypes.PlayAddressItem{
			Name: constants.VideoPlayAddressTypeWsFlv,
			Url:  fmt.Sprintf("%s://%s:%d/live/%s.flv", wsProtocol, msNode.IP, port, streamName),
		},
		Hls: &cTypes.PlayAddressItem{
			Name: constants.VideoPlayAddressTypeHls,
			Url:  fmt.Sprintf("%s://%s:%d/hls/ts/%s.m3u8", httpProtocol, msNode.IP, port, streamName),
		},
		HttpFmp4: &cTypes.PlayAddressItem{
			Name: constants.VideoPlayAddressTypeHttpFmp4,
			Url:  fmt.Sprintf("%s://%s:%d/m4s/live/%s.mp4", httpProtocol, msNode.IP, port, streamName),
		},
		Webrtc: &cTypes.PlayAddressItem{
			Name: constants.VideoPlayAddressTypeWebrtc,
			Url:  fmt.Sprintf("%s://%s:%d/webrtc/live/%s.whep", webrtcProtocol, msNode.IP, port, streamName),
		},
		WebrtcDC: &cTypes.PlayAddressItem{
			Name: constants.VideoPlayAddressTypeWebrtcDc,
			Url:  fmt.Sprintf("webrtc://%s:%d/webrtc/play/live/%s", msNode.IP, port, streamName),
		},
	}
	if msNode.IsDomain {
		if streamPlayProxyPath.WS != "" {
			data.WSFlv = &cTypes.PlayAddressItem{
				Name: constants.VideoPlayAddressTypeWsFlv,
				Url:  fmt.Sprintf("%s://%s/%s/live/%s.flv", wsProtocol, msNode.IP, streamPlayProxyPath.WS, streamName),
			}
		}

		if streamPlayProxyPath.HTTP != "" {
			data.HttpFlv = &cTypes.PlayAddressItem{
				Name: constants.VideoPlayAddressTypeHttpFlv,
				Url:  fmt.Sprintf("%s://%s/%s/live/%s.flv", httpProtocol, msNode.IP, streamPlayProxyPath.HTTP, streamName),
			}
			data.Hls = &cTypes.PlayAddressItem{
				Name: constants.VideoPlayAddressTypeHls,
				Url:  fmt.Sprintf("%s://%s/%s/hls/ts/%s.m3u8", httpProtocol, msNode.IP, streamPlayProxyPath.HTTP, streamName),
			}
			data.HttpFmp4 = &cTypes.PlayAddressItem{
				Name: constants.VideoPlayAddressTypeHttpFmp4,
				Url:  fmt.Sprintf("%s://%s/%s/m4s/live/%s.mp4", httpProtocol, msNode.IP, streamPlayProxyPath.HTTP, streamName),
			}

			data.Webrtc = &cTypes.PlayAddressItem{
				Name: constants.VideoPlayAddressTypeWebrtc,
				Url:  fmt.Sprintf("%s://%s/%s/webrtc/live/%s.whep", webrtcProtocol, msNode.IP, streamPlayProxyPath.HTTP, streamName),
			}
			data.WebrtcDC = &cTypes.PlayAddressItem{
				Name: constants.VideoPlayAddressTypeWebrtcDc,
				Url:  fmt.Sprintf("webrtc://%s/%s/webrtc/play/live/%s", msNode.IP, streamPlayProxyPath.HTTP, streamName),
			}
		}
	}

	return data
}
