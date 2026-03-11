// @Title        devices
// @Description  main
// @Create       yiyiyi 2025/7/19 15:57

package types

import (
	"skeyevss/core/repositories/models/channels"
	"skeyevss/core/repositories/models/devices"
)

type DeviceChannel struct {
	Device  *devices.Item
	Channel *channels.Item
}
type DeviceChannels struct {
	Devices  []*devices.Item
	Channels []*channels.Item
}

type DeviceOnlineStateResp struct {
	Channels map[string]uint `json:"channels"`
	Devices  map[string]uint `json:"devices"`
}

type (
	RtspStreamGroupItem struct {
		CId             uint64 `json:"cId"`
		ChannelUniqueId string `json:"channelUniqueId"`
		DeviceUniqueId  string `json:"deviceUniqueId"`
		StreamUrl       string `json:"streamUrl"`
	}

	RtspStreamGroupResp struct {
		Rtmp             []*RtspStreamGroupItem `json:"rtmp"`
		Onvif            []*RtspStreamGroupItem `json:"onvif"`
		StreamSourceRtsp []*RtspStreamGroupItem `json:"streamSourceRtsp"`
		Http             []*RtspStreamGroupItem `json:"http"`
	}

	DeviceStatisticsResp struct {
		ChannelOnlineCount  int64         `json:"channelOnlineCount"`
		ChannelOfflineCount int64         `json:"channelOfflineCount"`
		DeviceOnlineCount   int64         `json:"deviceOnlineCount"`
		DeviceOfflineCount  int64         `json:"deviceOfflineCount"`
		AccessProtocolGroup map[uint]uint `json:"accessProtocolGroup"`
	}

	ChannelMSRelItem struct {
		ChannelId       uint64   `json:"channelId"`
		ChannelUniqueId string   `json:"channelUniqueId"`
		DeviceUniqueId  string   `json:"deviceUniqueId"`
		MSIds           []uint64 `json:"msIds"`
	}

	RemoteReq struct {
		VssHttpTarget      string `json:"vssHttpTarget"`
		VssHttpUrl         string `json:"vssHttpUrl"`
		VssHttpUrlInternal string `json:"vssHttpUrlInternal"`
		VssSseTarget       string `json:"vssSseTarget"`
		VssSseUrl          string `json:"vssSseUrl"`
		Referer            string `json:"referer"`
	}
)
