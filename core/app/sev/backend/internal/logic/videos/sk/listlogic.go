package sk

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/app/sev/backend/internal/types"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
)

type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type ListItem struct {
	RecordPath string `json:"record_path"`
	Date       int64  `json:"date"`
	Duration   int64  `json:"duration"`
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListLogic) List(req *types.QueryVideoRecordsReq) (interface{}, *response.HttpErr) {
	videoRes, data, msAddress, err := records(l.ctx, l.svcCtx, req)
	if err != nil {
		return nil, err
	}

	if len(videoRes.Data) <= 0 {
		return &response.ListWithMapResp[[]*ListItem, string]{List: []*ListItem{}, Count: int64(0)}, nil
	}

	var times [][2]string
	for _, item := range videoRes.Data {
		item.RecordPath = msAddress + "/playback/" + strings.TrimPrefix(item.RecordPath, "/")
		times = append(times, [2]string{
			functions.NewTimer().FormatTimestamp(item.Date, ""),
			functions.NewTimer().FormatTimestamp(item.Date+item.Duration, ""),
		})
	}

	sort.Slice(
		videoRes.Data,
		func(i, j int) bool {
			var (
				start = time.UnixMilli(videoRes.Data[i].Date)
				end   = time.UnixMilli(videoRes.Data[j].Date)
			)
			if req.SortDate != nil && strings.ToLower(*req.SortDate) == strings.ToLower(string(orm.OrderAsc)) {
				return start.After(end)
			}

			return start.Before(end)
		},
	)

	var page = req.Page
	if page <= 0 {
		page = 1
	}

	var (
		totalCount = uint64(len(videoRes.Data))
		start      = (page - 1) * req.Limit
		end        = start + req.Limit
	)
	if start > totalCount {
		start = totalCount
	}

	if end > totalCount {
		end = totalCount
	}

	if start < end {
		return &response.ListWithMapResp[[]*ListItem, string]{
			Maps: map[string]interface{}{
				data.Channel.UniqueId: data.Channel,
			},
			List:  videoRes.Data[start:end],
			Count: int64(totalCount),
			Ext:   map[string]interface{}{"times": times},
		}, nil
	}

	return &response.ListWithMapResp[[]*ListItem, string]{
		Maps: map[string]interface{}{
			data.Channel.UniqueId: data.Channel,
		},
		List:  []*ListItem{},
		Count: int64(totalCount),
		Ext:   map[string]interface{}{"times": times},
	}, nil
}
