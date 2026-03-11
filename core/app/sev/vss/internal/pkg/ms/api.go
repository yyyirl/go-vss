package ms

import (
	"context"
	"errors"
	"fmt"
	"time"

	"skeyevss/core/app/sev/vss/internal/types"
	cTypes "skeyevss/core/common/types"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/functions/sc"
	"skeyevss/core/repositories/models/devices"
)

type MS struct {
	svcCtx   *types.ServiceContext
	ctx      context.Context
	reqHttps bool
}

func New(ctx context.Context, svcCtx *types.ServiceContext) *MS {
	return &MS{svcCtx: svcCtx, ctx: ctx}
}

type StreamInfo struct {
	StreamPort        uint
	TransportProtocol *devices.TransportProtocol
	MSNode            *cTypes.MSVoteNodeResp
	MediaServerUrl    string
}

func (l *MS) StreamInfo(deviceItem *devices.Item) (*StreamInfo, error) {
	var (
		streamPort        uint = 0
		transportProtocol      = deviceItem.TransportProtocol()
		msNode                 = l.VoteNode(deviceItem.MSIds)
	)
	if msNode == nil {
		return nil, errors.New("未设置流媒体源")
	}

	var mediaServerUrl = fmt.Sprintf("http://%s/api", msNode.Address)
	if l.svcCtx.Config.Sip.MediaServerStreamPortMax <= 0 || l.svcCtx.Config.Sip.MediaServerStreamPortMin <= 0 {
		return nil, errors.New("推流端口范围未设置")
	}

	if l.svcCtx.Config.Sip.MediaServerVssSameMachine {
		// media server和gbs在同一台机器  端口分配
		for i := l.svcCtx.Config.Sip.MediaServerStreamPortMin; i <= l.svcCtx.Config.Sip.MediaServerStreamPortMax; i++ {
			if sc.GetPid(int(i)) > 0 {
				continue
			}

			streamPort = i
			break
		}
	} else {
		// 不同机器 获取可用端口
		var err error
		streamPort, err = l.AvailablePort(mediaServerUrl, int(transportProtocol.Protocol))
		if err != nil {
			return nil, err
		}
	}

	return &StreamInfo{
		StreamPort:        streamPort,
		TransportProtocol: transportProtocol,
		MSNode:            msNode,
		MediaServerUrl:    mediaServerUrl,
	}, nil
}

func (l *MS) Snapshot(deviceItem *devices.Item, streamName string) ([]byte, error) {
	var msNode = l.VoteNode(deviceItem.MSIds)
	if msNode == nil {
		return nil, errors.New("未设置流媒体源")
	}

	var mediaServerUrl = fmt.Sprintf("http://%s/api", msNode.Address)
	if l.svcCtx.Config.Sip.MediaServerStreamPortMax <= 0 || l.svcCtx.Config.Sip.MediaServerStreamPortMin <= 0 {
		return nil, errors.New("推流端口范围未设置")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := functions.NewResty(ctx, &functions.RestyConfig{
		Mode: "pro",
	}).HttpGet(
		fmt.Sprintf("%s/stat/key_frame", mediaServerUrl),
		map[string]string{
			"stream_name": streamName,
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Body(), nil
}

// 向media server发送拉流请求
func (l *MS) RTPPub(req *types.SipVideoLiveInviteMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	type res struct {
		StreamName string `json:"stream_name"`
		SessionID  string `json:"session_id"`
		Port       uint   `json:"port"`
	}

	var resp cTypes.MediaResponse[res]
	if _, err := functions.NewResty(ctx, &functions.RestyConfig{
		Mode: l.svcCtx.Config.Mode,
	}).HttpPostJsonResJson(
		fmt.Sprintf("%s/ctrl/start_rtp_pub", req.MediaServerUrl),
		map[string]interface{}{
			"stream_name":                   req.StreamName,
			"timeout_ms":                    l.svcCtx.Config.Sip.MediaReceiveStreamTimeout * 1000,
			"is_tcp_flag":                   req.MediaProtocolMode,
			"is_wait_key_frame":             1,
			"is_tcp_active":                 req.MediaTransMode == "active",
			"auto_stop_pub_after_no_out_ms": l.svcCtx.Config.Sip.MediaNoWatchingTimeout * 1000,
			"speed":                         req.Speed,
			"download":                      req.Download,
		},
		&resp,
	); err != nil {
		return err
	}

	if resp.Code != 10000 {
		return fmt.Errorf("code: %d, message: %s", resp.Code, resp.Msg)
	}
	req.StreamPort = resp.Data.Port
	return nil
}

type StartRelyPullParams struct {
	StreamName               string
	StreamUrl                string
	AutoStopPullAfterNoOutMs int
	RtspMode                 uint
}

func (l *MS) StartRelyPull(address string, params *StartRelyPullParams) error {
	b, err := functions.NewRequest(l.svcCtx.Config.Mode).PostJson(
		fmt.Sprintf("http://%s/api/ctrl/start_relay_pull", address),
		map[string]interface{}{
			"stream_name":                    params.StreamName,
			"url":                            params.StreamUrl,
			"pull_retry_num":                 -1,
			"auto_stop_pull_after_no_out_ms": params.AutoStopPullAfterNoOutMs,
			"pull_timeout_ms":                15000,
			"rtsp_mode":                      params.RtspMode,
			"keep_live_form_get_parameter":   true,
		},
		false,
	)
	if err != nil {
		return err
	}

	type MSResp = struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			StreamName string `json:"stream_name"`
			SessionID  string `json:"session_id"`
		} `json:"data"`
	}
	var msResp MSResp
	if err := functions.JSONUnmarshal(b, &msResp); err != nil {
		return err
	}

	if msResp.Code != 10000 {
		return fmt.Errorf("response error: %s", msResp.Msg)
	}

	return nil
}

func (l *MS) StartRelayPush(address string, streamName, pushUrl string) error {
	b, err := functions.NewRequest(l.svcCtx.Config.Mode).PostJson(
		fmt.Sprintf("http://%s/api/ctrl/start_relay_push", address),
		map[string]interface{}{
			"stream_name":    streamName,
			"push_url":       pushUrl,
			"recon_interval": 5,
		},
		false,
	)
	if err != nil {
		return err
	}

	type MSResp = struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	var msResp MSResp
	if err := functions.JSONUnmarshal(b, &msResp); err != nil {
		return err
	}

	if msResp.Code != 10000 {
		return fmt.Errorf("response error: %s", msResp.Msg)
	}

	return nil
}

// 向media server发送拉流请求
func (l *MS) ACKRtpPub(req *types.SipVideoLiveInviteMessage, peerPort int, ip string, filesize uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	type res struct {
		StreamName string `json:"stream_name"`
		SessionID  string `json:"session_id"`
		Port       uint   `json:"port"`
	}
	var resp cTypes.MediaResponse[res]
	if _, err := functions.NewResty(ctx, &functions.RestyConfig{
		Mode: l.svcCtx.Config.Mode,
	}).HttpPostJsonResJson(
		fmt.Sprintf("%s/ctrl/ack_rtp_pub", req.MediaServerUrl),
		map[string]interface{}{
			"stream_name":   req.StreamName,
			"peer_port":     peerPort,
			"peer_ip":       ip,
			"is_tcp_flag":   req.MediaProtocolMode,
			"is_tcp_active": req.MediaTransMode == "active",
			"file_size":     filesize,
		},
		&resp,
	); err != nil {
		return err
	}

	if resp.Code != 10000 {
		return fmt.Errorf("code: %d, message: %s", resp.Code, resp.Msg)
	}

	return nil
}

// 获取可用端口
func (l *MS) AvailablePort(url string, protocol int) (uint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	type res struct {
		Port     uint `json:"port"`
		Protocol uint `json:"protocol"`
	}

	var resp cTypes.MediaResponse[res]
	if _, err := functions.NewResty(ctx, &functions.RestyConfig{
		Mode: l.svcCtx.Config.Mode,
	}).HttpPostJsonResJson(
		fmt.Sprintf("%s/stat/usable_port", url),
		map[string]interface{}{
			"max":      l.svcCtx.Config.Sip.MediaServerStreamPortMax,
			"min":      l.svcCtx.Config.Sip.MediaServerStreamPortMin,
			"protocol": protocol,
		},
		&resp,
	); err != nil {
		return 0, err
	}

	if resp.Code != 10000 {
		return 0, fmt.Errorf("code: %d, message: %s", resp.Code, resp.Msg)
	}

	return resp.Data.Port, nil
}

type StreamGroupItem struct {
	SessionID         string `json:"session_id"`
	Protocol          string `json:"protocol"`
	BaseType          string `json:"base_type"`
	RemoteAddr        string `json:"remote_addr"`
	StartTime         string `json:"start_time"`
	ReadBytesSum      uint   `json:"read_bytes_sum"`
	WroteBytesSum     uint   `json:"wrote_bytes_sum"`
	BitrateKbits      uint   `json:"bitrate_kbits"`
	ReadBitrateKbits  uint   `json:"read_bitrate_kbits"`
	WriteBitrateKbits uint   `json:"write_bitrate_kbits"`
}

type MSStartRtpPushResp struct {
	StreamName string `json:"stream_name"`
	SessionId  string `json:"session_id"`
	Port       int    `json:"port"`
}

type StreamGroupResp struct {
	StreamName  string `json:"stream_name"`
	AppName     string `json:"app_name"`
	AudioCodec  string `json:"audio_codec"`
	VideoCodec  string `json:"video_codec"`
	VideoWidth  uint   `json:"video_width"`
	VideoHeight uint   `json:"video_height"`

	// 向ms推流(向ms入) 判断 (国标, rtmp push) 设备主动向ms推流
	Pub *StreamGroupItem `json:"pub"`
	// ms主动拉流(向ms入)
	Pull struct {
		SessionID         string `json:"session_id"`
		Protocol          string `json:"protocol"`
		BaseType          string `json:"base_type"`
		RemoteAddr        string `json:"remote_addr"`
		StartTime         string `json:"start_time"`
		ReadBytesSum      uint   `json:"read_bytes_sum"`
		WroteBytesSum     uint   `json:"wrote_bytes_sum"`
		BitrateKbits      uint   `json:"bitrate_kbits"`
		ReadBitrateKbits  uint   `json:"read_bitrate_kbits"`
		WriteBitrateKbits uint   `json:"write_bitrate_kbits"`
	} `json:"pull"`

	// 是否有人播放观看(ms 输出) 有几人就有几个元素
	Subs []*StreamGroupItem `json:"subs"`
	// 设备推流给ms(ms 输出) ms再推给指定服务端(转推)
	Pushs []*StreamGroupItem `json:"pushs"`
}

type QueryRecordByNamesItem struct {
	RecordPath string `json:"record_path"`
	StreamName string `json:"stream_name"`
}

// 获取流信息
func (l *MS) GetStreamGroup(url, streamName string) (*StreamGroupResp, uint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var resp cTypes.MediaResponse[StreamGroupResp]
	if _, err := functions.NewResty(ctx, &functions.RestyConfig{
		Mode: l.svcCtx.Config.Mode,
	}).HttpGetResJson(
		fmt.Sprintf("%s/stat/group", url),
		map[string]string{
			"stream_name": streamName,
		},
		&resp,
	); err != nil {
		return nil, resp.Code, err
	}

	if resp.Code != 10000 {
		return nil, resp.Code, fmt.Errorf("code: %d, message: %s", resp.Code, resp.Msg)
	}

	return &resp.Data, resp.Code, nil
}

// 获取流分组信息
func (l *MS) GetStreamAllGroup(url string) ([]*StreamGroupResp, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var resp cTypes.MediaResponse[[]*StreamGroupResp]
	if _, err := functions.NewResty(ctx, &functions.RestyConfig{
		Mode: l.svcCtx.Config.Mode,
	}).HttpGetResJson(
		fmt.Sprintf("http://%s/api/stat/all_group", url),
		nil,
		&resp,
	); err != nil {
		return nil, err
	}

	if resp.Code != 10000 {
		return nil, fmt.Errorf("code: %d, message: %s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}

// 按录像流名称列表查询服务录像
func (l *MS) QueryRecordByNames(url string, req types.MsQueryRecordByNameReq) ([]*QueryRecordByNamesItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var resp cTypes.MediaResponse[[]*QueryRecordByNamesItem]
	if _, err := functions.NewResty(ctx, &functions.RestyConfig{
		Mode: l.svcCtx.Config.Mode,
	}).HttpPostJsonResJson(
		fmt.Sprintf("http://%s/api/record/query_by_names", url),
		map[string]interface{}{
			"stream_names": req.StreamNames,
			"record_type":  req.RecordType,
		},
		&resp,
	); err != nil {
		return nil, err
	}

	if resp.Code != 10000 {
		return nil, fmt.Errorf("code: %d, message: %s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}

// 踢出会话
func (l *MS) KickSession(url, streamName, sessionId string) (*StreamGroupResp, uint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var resp cTypes.MediaResponse[StreamGroupResp]
	if _, err := functions.NewResty(
		ctx, &functions.RestyConfig{
			Mode: l.svcCtx.Config.Mode,
		},
	).HttpPostJsonResJson(
		fmt.Sprintf("%s/ctrl/kick_session", url),
		map[string]interface{}{
			"stream_name": streamName,
			"session_id":  sessionId,
		},
		&resp,
	); err != nil {
		return nil, resp.Code, err
	}

	if resp.Code != 10000 {
		return nil, resp.Code, fmt.Errorf("code: %d, message: %s", resp.Code, resp.Msg)
	}

	return &resp.Data, resp.Code, nil
}

// 获取流媒体端口
func (l *MS) GetMSConf(url string) (*cTypes.MSConfResp, uint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var resp cTypes.MediaResponse[cTypes.MSConfResp]
	if _, err := functions.NewResty(ctx, &functions.RestyConfig{
		Mode: l.svcCtx.Config.Mode,
	}).HttpGetResJson(
		fmt.Sprintf("%s/api/config/sms_config_brief", url),
		nil,
		&resp,
	); err != nil {
		return nil, resp.Code, err
	}

	if resp.Code != 10000 {
		return nil, resp.Code, fmt.Errorf("code: %d, message: %s", resp.Code, resp.Msg)
	}

	return &resp.Data, resp.Code, nil
}

// 获取流媒体端口
func (l *MS) GetMSConf1(url string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var resp cTypes.MediaResponse[map[string]interface{}]
	if _, err := functions.NewResty(ctx, &functions.RestyConfig{
		Mode: l.svcCtx.Config.Mode,
	}).HttpGetResJson(
		fmt.Sprintf("%s/api/config/sms_config_brief", url),
		nil,
		&resp,
	); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (l *MS) Reload(url string, data types.MsReloadReq) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var resp cTypes.MediaResponseBase
	if _, err := functions.NewResty(ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/ctrl/sms_reload", url),
		map[string]interface{}{
			"config": data.Config,
			"reboot": data.Reboot,
			"delay":  data.Delay,
		},
		&resp,
	); err != nil {
		return err
	}

	if resp.Code != 10000 {
		return fmt.Errorf("code: %d, message: %s", resp.Code, resp.Msg)
	}

	return nil
}

func (l *MS) StartRtpPush(url string, data map[string]interface{}) (*MSStartRtpPushResp, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var resp cTypes.MediaResponse[MSStartRtpPushResp]
	if _, err := functions.NewResty(ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/ctrl/start_rtp_push", url),
		data,
		&resp,
	); err != nil {
		return nil, err
	}

	if resp.Code != 10000 {
		return nil, fmt.Errorf("code: %d, message: %s", resp.Code, resp.Msg)
	}

	return &resp.Data, nil
}

func (l *MS) StopRtpPush(url string, streamName, sessionId string) (uint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var resp cTypes.MediaResponseBase
	if _, err := functions.NewResty(ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/ctrl/stop_rtp_push", url),
		map[string]interface{}{
			"stream_name": streamName,
			"session_id":  sessionId,
		},
		&resp,
	); err != nil {
		return 400, err
	}

	return resp.Code, nil
}

func (l *MS) AckRtpPush(url string, streamName, sessionId string) (uint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var resp cTypes.MediaResponseBase
	if _, err := functions.NewResty(ctx, &functions.RestyConfig{Mode: l.svcCtx.Config.Mode}).HttpPostJsonResJson(
		fmt.Sprintf("%s/api/ctrl/ack_rtp_push", url),
		map[string]interface{}{
			"stream_name": streamName,
			"session_id":  sessionId,
		},
		&resp,
	); err != nil {
		return 400, err
	}

	return resp.Code, nil
}

// 停止流
func (l *MS) StopMSStream(msAddress, streamName, Type string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 停止回源拉流 https://gitee.com/cdu_yolin/skeyesms/blob/master/document/api/general_api.md
	var (
		resp map[string]interface{}
		url  = fmt.Sprintf("http://%s/api/ctrl/stop_relay_pull?stream_name=%s", msAddress, streamName)
	)
	if Type == "pub" {
		url = fmt.Sprintf("http://%s/api/ctrl/stop_rtp_pub?stream_name=%s", msAddress, streamName)
	}

	if _, err := functions.NewResty(
		ctx, &functions.RestyConfig{
			Mode: l.svcCtx.Config.Mode,
		},
	).HttpPostJsonResJson(url, map[string]interface{}{}, &resp); err != nil {
		return err
	}

	return nil
}

// 批量停止流
func (l *MS) StopMultiMSStream(msAddress, streamName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var resp map[string]interface{}
	_, err := functions.NewResty(
		ctx, &functions.RestyConfig{
			Mode: l.svcCtx.Config.Mode,
		},
	).HttpPostJsonResJson(
		fmt.Sprintf("http://%s/api/ctrl/stop_idle_rtp_pubs?stream_name_x=%s", msAddress, streamName),
		map[string]interface{}{"stream_name_x": streamName},
		&resp,
	)
	return err
}

func (l *MS) WithHttps(v bool) *MS {
	l.reqHttps = v
	return l
}

func (l *MS) VoteNode(ids []uint64) *cTypes.MSVoteNodeResp {
	var ip = l.svcCtx.Config.InternalIp
	if l.svcCtx.Config.Sip.UseExternalWan {
		ip = l.svcCtx.Config.ExternalIp
	}

	if l.ctx != nil && contextx.GetIsInternalReq(l.ctx) {
		ip = l.svcCtx.Config.InternalIp
	}

	// 默认节点
	var (
		httpPort  = uint(l.svcCtx.Config.SevBase.MediaServerPort)
		httpsPort = uint(l.svcCtx.Config.SevBase.MediaServerHttpsPort)
		rtspPort  = uint(l.svcCtx.Config.SevBase.MediaServerRtspPort)
		rtmpPort  = uint(l.svcCtx.Config.SevBase.MediaServerRtmpPort)
		node      = fmt.Sprintf("%s:%d", ip, httpPort)
		def       = &cTypes.MSVoteNodeResp{
			Address: node,
			Name:    "default",
			IsDef:   true,

			InternalIP:   l.svcCtx.Config.InternalIp,
			ExtIP:        l.svcCtx.Config.ExternalIp,
			IP:           ip,
			HttpPort:     httpPort,
			HttpsPort:    httpsPort,
			RtspPort:     rtspPort,
			RtmpPort:     rtmpPort,
			UseHttpsPlay: l.reqHttps,
		}
	)
	// 使用配置域名
	if l.svcCtx.Config.Domain != "" {
		def.IP = l.svcCtx.Config.Domain
		def.IsDomain = true
	}

	if len(ids) <= 0 {
		return def
	}

	var nodes []*cTypes.MSVoteNodeResp
	for _, item := range l.svcCtx.MediaServerRecords {
		if functions.Contains(item.ID, ids) && item.IP != "" && item.Port > 0 {
			var ip = item.IP
			if l.svcCtx.Config.Sip.UseExternalWan {
				ip = item.ExtIP
			}

			if l.ctx != nil && contextx.GetIsInternalReq(l.ctx) {
				ip = item.IP
			}

			nodes = append(nodes, &cTypes.MSVoteNodeResp{
				Address:    fmt.Sprintf("%s:%d", item.IP, item.Port),
				Name:       item.Name,
				ID:         item.ID,
				IP:         ip,
				InternalIP: item.IP,
				ExtIP:      item.ExtIP,
				HttpPort:   item.Port,
				// UseHttpsPlay: setting.Content().MediaServerUseHttpsPlay,
			})
		}
	}

	if len(nodes) <= 0 {
		return def
	}

	return nodes[0]
	// return nodes[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(nodes))]
}

func (l *MS) VoteNodeItem(msIP string) uint64 {
	// 默认节点
	var ip = l.svcCtx.Config.InternalIp
	if msIP == "" || msIP == ip {
		return 0
	}

	for _, item := range l.svcCtx.MediaServerRecords {
		if msIP == item.IP {
			return item.ID
		}
	}

	return 0
}
