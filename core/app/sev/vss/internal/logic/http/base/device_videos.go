// @Title        设备录像
// @Description  main
// @Create       yiyiyi 2025/7/23 08:55

package base

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/vss/internal/pkg/sip"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/constants"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/dt"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

var (
	_            types.HttpRHandleLogic[*VideosLogic, types.QueryVideoRecordsReq] = (*VideosLogic)(nil)
	VVideosLogic                                                                  = new(VideosLogic)
)

type VideosLogic struct {
	ctx    context.Context
	c      *gin.Context
	svcCtx *types.ServiceContext
}

func (l *VideosLogic) New(ctx context.Context, c *gin.Context, svcCtx *types.ServiceContext) *VideosLogic {
	return &VideosLogic{
		ctx:    ctx,
		c:      c,
		svcCtx: svcCtx,
	}
}

func (l *VideosLogic) Path() string {
	return "/device-videos"
}

func (l *VideosLogic) DO(req types.QueryVideoRecordsReq) *types.HttpResponse {
	var (
		err  error
		date = strings.ReplaceAll(
			strings.Join(
				strings.Split(
					strings.ReplaceAll(
						functions.NewTimer().FormatTimestamp(req.Day, functions.TimeFormatYmdhis),
						"-",
						"",
					),
					"",
				)[2:],
				"",
			),
			"0",
			"",
		)
	)
	req.SN, err = strconv.ParseInt(date, 10, 64)
	if err != nil {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0010),
		}
	}

	var (
		fetchDeviceVideoStateKey = fmt.Sprintf("%s-%s-%d", req.DeviceUniqueId, req.ChannelUniqueId, req.Day)
		cacheKey                 = sip.VideoRecordMapKey(req.DeviceUniqueId, req.ChannelUniqueId, req.SN)
	)
	// 重复请求 获取上一次请求缓存
	if l.svcCtx.FetchDeviceVideoState.Contains(fetchDeviceVideoStateKey) {
		var (
			ticker  = time.NewTicker(100 * time.Millisecond)
			timeout = time.After(5 * time.Second)
		)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if res := l.fetchData(req, cacheKey); res != nil {
					return res
				}

			case <-timeout:
				return &types.HttpResponse{
					Err: response.MakeError(response.NewHttpRespMessage().Str("设备视频获取超时"), localization.M0010),
				}
			}
		}
	}

	l.svcCtx.FetchDeviceVideoState.Add(fetchDeviceVideoStateKey)
	defer func() {
		dt.SetTimeout(
			500*time.Millisecond,
			func() {
				l.svcCtx.FetchDeviceVideoState.Remove(fetchDeviceVideoStateKey)
			},
		)
	}()

	if _, ok := l.svcCtx.SipCatalogLoopMap.Get(req.DeviceUniqueId); !ok {
		return &types.HttpResponse{
			Err: response.MakeError(response.NewHttpRespMessage().Err(constants.DeviceUnregistered), localization.M00300),
		}
	}

	// 向设备发送获取视频记录请求
	l.svcCtx.SipSendQueryVideoRecords <- &req

	var (
		ticker  = time.NewTicker(100 * time.Millisecond)
		timeout = time.After(5 * time.Second)
	)
	defer func() {
		ticker.Stop()

		dt.SetTimeout(
			500*time.Millisecond,
			func() {
				l.svcCtx.SipMessageVideoRecordMap.Remove(cacheKey)
			},
		)
	}()

	for {
		select {
		case <-ticker.C:
			var res = l.fetchData(req, cacheKey)
			if res != nil {
				return res
			}
			continue

		case <-timeout:
			return &types.HttpResponse{
				Err: response.MakeError(response.NewHttpRespMessage().Str("设备视频获取超时"), localization.M0010),
			}
		}
	}
}

func (l *VideosLogic) fetchData(req types.QueryVideoRecordsReq, cacheKey string) *types.HttpResponse {
	data, ok := l.svcCtx.SipMessageVideoRecordMap.Get(cacheKey)
	if !ok || len(data.List) < data.Total {
		return nil
	}

	var times [][2]string
	for _, item := range data.List {
		item.StartTime = strings.ReplaceAll(item.StartTime, "T", " ")
		item.EndTime = strings.ReplaceAll(item.EndTime, "T", " ")
		times = append(times, [2]string{item.StartTime, item.EndTime})
	}

	var list = functions.PickWithPageOffset(int(req.Page), int(req.Limit), data.List)
	for _, item := range list {
		item.UniqueId = strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(
					item.StartTime+item.EndTime, "-", "",
				), " ", "",
			), ":", "",
		)

	}

	return &types.HttpResponse{Data: map[string]interface{}{"list": list, "count": data.Total, "slices": times}}
}
