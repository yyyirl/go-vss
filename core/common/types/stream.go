// @Title        a
// @Description  main
// @Create       yiyiyi 2025/9/3 14:50

package types

type (
	OnvifDeviceItem struct {
		UUID            string   `json:"uuid"`
		OriginalUid     string   `json:"originalUid"`
		Name            string   `json:"name"`
		Address         string   `json:"address"`
		ServiceURLs     []string `json:"service_urls"`
		Types           []string `json:"types"`
		XAddrs          []string `json:"xaddrs"`
		Scopes          []string `json:"scopes"`
		Model           string   `json:"model"`
		Manufacturer    string   `json:"manufacturer"`
		FirmwareVersion string   `json:"firmware_version"`
		SerialNumber    string   `json:"serial_number"`
	}

	OnvifDeviceProfileItem struct {
		Profile      string `json:"name"`
		ProfileToken string `json:"token"`
		Url          string `json:"url"`
	}

	StreamResp struct {
		AccessProtocolName string `json:"accessProtocolName"`
		AccessProtocol     uint   `json:"accessProtocol"`
		MediaServerID      uint64 `json:"mediaServerID"`
		MediaServerNode    string `json:"mediaServerNode"`
		MediaServerName    string `json:"mediaServerName"`
		DeviceID           string `json:"deviceID"`
		ChannelID          string `json:"channelID"`
		StreamName         string `json:"streamName"`
		ChannelName        string `json:"channelName"`
		DeviceName         string `json:"deviceName"`
		StreamUrl          string `json:"streamUrl"`
		// Http            string `json:"http"`
		// Webrtc          string `json:"webrtc"`

		ChannelOnlineState bool         `json:"channelOnlineState"`
		Addresses          *PlayAddress `json:"addresses"`

		StartAt string `json:"startAt,omitempty"`
		EndAt   string `json:"endAt,omitempty"`
	}
)

type MSConfResp struct {
	ConfVersion                 string `json:"conf_version"`
	CheckSessionDisposeInterval uint   `json:"check_session_dispose_interval"`
	UpdateSessionStateInterval  uint   `json:"update_session_state_interval"`
	ManagerChanSize             uint   `json:"manager_chan_size"`
	AdjustPts                   bool   `json:"adjust_pts"`
	MaxOpenFiles                uint   `json:"max_open_files"`

	LogIsToStdout bool `json:"log_is_to_stdout"`

	HttpListenPort  uint   `json:"http_listen_port"`
	HttpsListenPort uint   `json:"https_listen_port"`
	HttpsCertFile   string `json:"https_cert_file"`
	HttpsKeyFile    string `json:"https_key_file"`

	RtmpEnable         bool `json:"rtmp_enable"`
	RtmpPort           uint `json:"rtmp_port"`
	RtmpsEnable        bool `json:"rtmps_enable"`
	RtmpsPort          uint `json:"rtmps_port"`
	RtmpOverQuicEnable bool `json:"rtmp_over_quic_enable"`
	RtmpOverQuicPort   uint `json:"rtmp_over_quic_port"`
	RtmpOverKcpEnable  bool `json:"rtmp_over_kcp_enable"`
	RtmpOverKcpPort    uint `json:"rtmp_over_kcp_port"`

	RtspEnable   bool `json:"rtsp_enable"`
	RtspPort     uint `json:"rtsp_port"`
	RtspsEnable  bool `json:"rtsps_enable"`
	RtspsPort    uint `json:"rtsps_port"`
	WsRtspEnable bool `json:"ws_rtsp_enable"`
	WsRtspPort   uint `json:"ws_rtsp_port"`

	RtcEnable          bool     `json:"rtc_enable"`
	RtcEnableHttps     bool     `json:"rtc_enable_https"`
	RtcIceUdpMinPort   uint     `json:"rtc_iceUdpMinPort"`
	RtcIceUdpMaxPort   uint     `json:"rtc_iceUdpMaxPort"`
	RtcIceTcpMinPort   uint     `json:"rtc_iceTcpMinPort"`
	RtcIceTcpMaxPort   uint     `json:"rtc_iceTcpMaxPort"`
	RtcIceHostNatToIps []string `json:"rtc_iceHostNatToIps"`

	// PPROF Config
	PprofEnable bool `json:"pprof_enable"`

	// http notify Config
	HttpNotifyEnable            bool   `json:"http_notify_enable"`
	HttpNotifyUpdateIntervalSec uint   `json:"http_notify_update_interval_sec"`
	OnUpdate                    string `json:"on_update"`
	OnPubStart                  string `json:"on_pub_start"`
	OnPubStop                   string `json:"on_pub_stop"`
	OnPushStart                 string `json:"on_push_start"`
	OnPushStop                  string `json:"on_push_stop"`
	OnSubStart                  string `json:"on_sub_start"`
	OnSubStop                   string `json:"on_sub_stop"`
	OnRelayPullStart            string `json:"on_relay_pull_start"`
	OnRelayPullStop             string `json:"on_relay_pull_stop"`
	OnRtmpConnect               string `json:"on_rtmp_connect"`
	OnServerStart               string `json:"on_server_start"`
	OnHlsMakeTs                 string `json:"on_hls_make_ts"`
	OnReportStat                string `json:"on_report_stat"`
	OnReportFrameInfo           string `json:"on_report_frame_info"`

	RecordEnableFlv            bool   `json:"record_enable_flv"`
	RecordFlvOutPath           string `json:"record_flv_out_path"`
	RecordEnableMpegts         bool   `json:"record_enable_mpegts"`
	RecordMpegtsOutPath        string `json:"record_mpegts_out_path"`
	RecordEnableFmp4           bool   `json:"record_enable_fmp4"`
	RecordFmp4OutPath          string `json:"record_fmp4_out_path"`
	RecordEnableRecordInterval bool   `json:"record_enable_record_interval"`
	RecordRecordInterval       uint   `json:"record_record_interval"`

	PlaybackEnable     bool   `json:"playback_enable"`
	PlaybackUrlPattern string `json:"playback_url_pattern"`

	HlsEnable      bool   `json:"hls_enable"`
	HlsOutPath     string `json:"hls_out_path"`
	HlsCleanupMode uint   `json:"hls_cleanup_mode"`
}

type MediaResponse[T any] struct {
	Error string `json:"error"`
	Code  uint   `json:"code"`
	Msg   string `json:"msg"`
	Data  T      `json:"data"`
}

type MediaResponseBase struct {
	Error string `json:"error"`
	Code  uint   `json:"code"`
	Msg   string `json:"msg"`
}

type MSVoteNodeResp struct {
	Address string
	Name    string
	ID      uint64
	IsDef   bool

	InternalIP,
	ExtIP,
	IP string
	HttpPort,
	HttpsPort,
	RtspPort,
	RtmpPort uint

	UseHttpsPlay,
	IsDomain bool
}

type (
	PlayAddressItem struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}

	PlayAddress struct {
		Rtsp     *PlayAddressItem `json:"rtsp"`
		Rtmp     *PlayAddressItem `json:"rtmp"`
		HttpFlv  *PlayAddressItem `json:"httpFlv"`
		Hls      *PlayAddressItem `json:"hls"`
		HttpFmp4 *PlayAddressItem `json:"httpFmp4"`
		Webrtc   *PlayAddressItem `json:"webrtc"`
		WebrtcDC *PlayAddressItem `json:"webrtcDC"`
		WSFlv    *PlayAddressItem `json:"wsFlv"`
	}
)
