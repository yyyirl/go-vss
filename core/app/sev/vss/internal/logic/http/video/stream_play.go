package video

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/vss/internal/logic/http/gbs"
	"skeyevss/core/app/sev/vss/internal/pkg/ms"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/common/stream"
	ctypes "skeyevss/core/common/types"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/ff"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/devices"
)

var (
	_ types.HttpRHandleLogic[*StreamPlayLogic, types.VideoStreamReq] = (*StreamPlayLogic)(nil)

	VStreamPlayLogic = new(StreamPlayLogic)
)

type StreamPlayLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *StreamPlayLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *StreamPlayLogic {
	return &StreamPlayLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *StreamPlayLogic) Path() string {
	return "/video/stream"
}

func (l *StreamPlayLogic) DO(req types.VideoStreamReq) *types.HttpResponse {
	if req.DeviceUniqueId == "" || req.ChannelUniqueId == "" {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("参数错误"), localization.M0001),
		}
	}

	res, err := response.NewRpcToHttpResp[*deviceservice.Response, *ctypes.DeviceChannel]().Parse(
		func() (*deviceservice.Response, error) {
			return l.svcCtx.RpcClients.Device.DeviceChannel(l.ctx, &deviceservice.DeviceChannelReq{
				ChannelUniqueId: req.ChannelUniqueId,
				DeviceUniqueId:  req.DeviceUniqueId,
			})
		},
	)
	if err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str(err.Error), localization.M0010),
		}
	}

	if res.Data.Device == nil || res.Data.Channel == nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("设备获取失败"), localization.M0010),
		}
	}

	var (
		autoStopPullAfterNoOutMs      = 60000
		playType                      = stream.PlayTypePlay
		transportProtocol             = res.Data.Device.TransportProtocol()
		rtspMode                 uint = 0 // 0 tcp 1 udp
	)
	// 回放
	if req.EndAt > 0 && req.StartAt > 0 {
		playType = stream.PlayTypePlayback
		autoStopPullAfterNoOutMs = 10000
	}

	var (
		msNode     = ms.New(l.ctx, l.svcCtx).WithHttps(req.Https).VoteNode(res.Data.Device.MSIds)
		streamUrl  string
		streamName = stream.New().Produce(req.DeviceUniqueId, req.ChannelUniqueId, playType)
	)

	if l.c != nil {
		if v := l.c.Query("streamName"); v != "" {
			streamName = v
		}
	}

	if msNode == nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("未设置流媒体源"), localization.M0010),
		}
	}

	var (
		channelName = res.Data.Channel.Label
		deviceName  = res.Data.Device.Label
	)
	if channelName == "" {
		channelName = res.Data.Channel.Name
	}

	if deviceName == "" {
		deviceName = res.Data.Device.Name
	}

	var data = &ctypes.StreamResp{
		AccessProtocolName: devices.AccessProtocols[res.Data.Device.AccessProtocol],
		AccessProtocol:     res.Data.Device.AccessProtocol,
		MediaServerID:      msNode.ID,
		MediaServerNode:    msNode.Address,
		MediaServerName:    msNode.Name,
		DeviceID:           res.Data.Channel.DeviceUniqueId,
		ChannelID:          res.Data.Channel.UniqueId,
		StreamUrl:          streamUrl,
		StreamName:         streamName,
		ChannelName:        channelName,
		DeviceName:         deviceName,
		ChannelOnlineState: res.Data.Channel.Online == 1,
	}
	if !msNode.IsDef {
		// 获取端口
		resp, _, err := ms.New(l.ctx, l.svcCtx).GetMSConf(fmt.Sprintf("http://%s", msNode.Address))
		if err != nil {
			return &types.HttpResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010),
			}
		}

		data.Addresses = stream.New().PlayAddress(
			l.svcCtx.Config.StreamPlayProxyPath,
			&ctypes.MSVoteNodeResp{
				Address: msNode.Address,
				Name:    msNode.Name,
				ID:      msNode.ID,
				IP:      msNode.IP,

				HttpPort:     resp.HttpListenPort,
				HttpsPort:    resp.HttpsListenPort,
				RtspPort:     resp.RtspPort,
				RtmpPort:     resp.RtmpPort,
				UseHttpsPlay: msNode.UseHttpsPlay,
			},
			streamName,
		)
	} else {
		data.Addresses = stream.New().PlayAddress(l.svcCtx.Config.StreamPlayProxyPath, msNode, streamName)
	}

	if req.StartAt > 0 {
		data.StartAt = functions.NewTimer().FormatTimestamp(req.StartAt, "")
	}

	if req.EndAt > 0 {
		data.EndAt = functions.NewTimer().FormatTimestamp(req.EndAt, "")
	}

	switch res.Data.Device.AccessProtocol { // 设备接入协议
	case devices.AccessProtocol_1: // 流媒体源 // 通知 拉流 on_relay_pull_start
		streamUrl = res.Data.Device.StreamUrl
		if transportProtocol.MediaProtocolMode == 0 {
			rtspMode = 1
		}

	case devices.AccessProtocol_2: // RTMP推流 // 通知 推流 on_pub_start
		// RTMP推流不需要调用start_relay_pull
		// streamUrl = fmt.Sprintf("%s/%s", strings.Trim(res.Data.Device.StreamUrl, "/"), streamName)

	case devices.AccessProtocol_3: // ONVIF协议 // 通知 拉流 on_relay_pull_start
		streamUrl = res.Data.Channel.StreamUrl
		if transportProtocol.MediaProtocolMode == 0 {
			rtspMode = 1
		}

	case devices.AccessProtocol_4: // GB28181协议 // 通知 推流 on_pub_start
		// 回放关闭流
		if playType == stream.PlayTypePlayback {
			if err := ms.New(l.ctx, l.svcCtx).StopMultiMSStream(
				msNode.Address, fmt.Sprintf("%s%s*", strings.Split(streamName, string(playType))[0], playType),
			); err != nil {
				functions.LogError("回放流停止失败, err:", err)
			} else {
				time.Sleep(1 * time.Second)
			}
		}

		// 发送invite
		if res := gbs.InviteLogic.New(l.ctx, l.c, l.svcCtx).Invite(&gbs.InviteParams{
			DeviceUniqueId: res.Data.Channel.DeviceUniqueId,
			ChannelID:      res.Data.Channel.UniqueId,
			PlayType:       playType,
			StartAt:        data.StartAt,
			EndAt:          data.EndAt,
			DeviceItem:     res.Data.Device,
			ChannelItem:    res.Data.Channel,
			Download:       req.Download,
			Speed:          req.Speed,
			StreamName:     streamName,
			Caller:         "http 请求stream play invite",
		}); res != nil && res.Err != nil {
			return res
		}

		go l.saveChannelSnapshot(streamName, res.Data)
		return &types.HttpResponse{Data: data}

	case devices.AccessProtocol_5: // EHOME协议 // 通知 推流 on_pub_start
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str("未匹配的类型: EHOME协议"), localization.M0010),
		}

	default:
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Str(fmt.Sprintf("未匹配的协议类型: %d", res.Data.Device.AccessProtocol)), localization.M0010),
		}
	}

	if streamUrl != "" {
		if err := ms.New(l.ctx, l.svcCtx).StartRelyPull(msNode.Address, &ms.StartRelyPullParams{
			StreamName:               streamName,
			StreamUrl:                streamUrl,
			AutoStopPullAfterNoOutMs: autoStopPullAfterNoOutMs,
			RtspMode:                 rtspMode,
		}); err != nil {
			return &types.HttpResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010),
			}
		}
	}

	go l.saveChannelSnapshot(streamName, res.Data)

	if res.Data.Channel.CdnState == 1 && res.Data.Channel.CdnUrl != "" && strings.HasPrefix(strings.ToLower(res.Data.Channel.CdnUrl), "rtmp") {
		go func() {
			if err := ms.New(l.ctx, l.svcCtx).StartRelayPush(msNode.Address, streamName, res.Data.Channel.CdnUrl); err != nil {
				functions.LogError("start_relay_push 请求失败", err.Error())
			}
		}()
	}

	return &types.HttpResponse{Data: data}
}

func (l *StreamPlayLogic) saveChannelSnapshot(streamName string, res *ctypes.DeviceChannel) {
	defer func() {
		if err := recover(); err != nil {
			functions.LogError("设置快照截图失败, panic err: ", err)
		}
	}()

	// 获取快照
	bytes, err := ms.New(l.ctx, l.svcCtx).Snapshot(res.Device, streamName)
	if err != nil {
		functions.LogError("snapshot获取失败, err: ", err)
		return
	}

	tmpSnapshotSaveFileAbsPath, err := filepath.Abs(path.Join("tmp", fmt.Sprintf("%s.jpg", functions.UniqueId())))
	if err != nil {
		functions.LogError("I帧文件绝对路径获取失败[2]")
		return
	}

	snapshotTmpFileAbsPath, err := filepath.Abs(path.Join("tmp", fmt.Sprintf("%s.raw", functions.UniqueId())))
	if err != nil {
		functions.LogError("I帧文件绝对路径获取失败[0]")
		return
	}

	defer func() {
		_ = os.Remove(tmpSnapshotSaveFileAbsPath)
		_ = os.Remove(snapshotTmpFileAbsPath)
	}()

	if err := functions.WriteToFile(snapshotTmpFileAbsPath, string(bytes)); err != nil {
		functions.LogError(fmt.Sprintf("I帧文件写入[%s]失败,err： %s", snapshotTmpFileAbsPath, err.Error()))
		return
	}

	if err := functions.WriteToFile(tmpSnapshotSaveFileAbsPath, ""); err != nil {
		functions.LogError("I帧文件写入失败[1]")
		return
	}

	var snapshotSaveFilePath = stream.New().Snapshot(l.svcCtx.Config.SaveVideoSnapshotDir, res.Channel.DeviceUniqueId, res.Channel.UniqueId)
	saveVideoSnapshotFile, err := filepath.Abs(snapshotSaveFilePath)
	if err != nil {
		functions.LogError("I帧文件绝对路径获取失败[1]")
		return
	}

	// 保存快照文件
	if err := ff.NewFFMpeg(l.svcCtx.Config.FFMpeg).SnapFile(snapshotTmpFileAbsPath, tmpSnapshotSaveFileAbsPath); err != nil {
		functions.LogError("快照文件保存失败[1], err:", err)
		return
	}

	if err := functions.MakeDir(l.svcCtx.Config.SaveVideoSnapshotDir); err != nil {
		functions.LogError("快照目录创建失败")
		return
	}

	if err := functions.Mv(tmpSnapshotSaveFileAbsPath, saveVideoSnapshotFile); err != nil {
		functions.LogError("快照文件保存失败[2], err:", err)
	}
}
