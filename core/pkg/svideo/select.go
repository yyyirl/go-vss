// @Title        select
// @Description  main
// @Create       yiyiyi 2025/8/18 11:55

package svideo

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
)

func (s *SVideo) Select(params *Params) (*response.ListWithMapResp[[]*Item, string], error) {
	if !functions.IsDir(s.SaveDir) {
		return &response.ListWithMapResp[[]*Item, string]{List: []*Item{}, Count: 0}, nil
	}

	var (
		records []*Item
		index   uint64 = 1
	)
	if err := filepath.Walk(s.SaveDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(path) != ".mp4" {
			return nil
		}

		var item = s.Parse(path, index)
		if item == nil {
			return nil
		}

		if params.DeviceUniqueId != "" && item.DeviceUniqueId != params.DeviceUniqueId {
			return nil
		}

		if params.ChannelUniqueId != "" && item.ChannelUniqueId != params.ChannelUniqueId {
			return nil
		}

		if params.Date != nil {
			var sart = uint64(item.StartTime.UnixMilli())
			if sart <= params.Date.Start || sart >= params.Date.End {
				return nil
			}
		}

		if params.Urgent != nil {
			if *params.Urgent != item.Urgent {
				return nil
			}
		}

		records = append(records, item)
		index += 1
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk directory: %v", err)
	}

	if params.UrgentDesc != nil {
		sort.Slice(records, func(i, j int) bool {
			if records[i].Urgent != records[j].Urgent {
				if *params.UrgentDesc {
					return !records[i].Urgent
				} else {
					return records[i].Urgent
				}
			}

			return records[i].Urgent
		})
	} else {
		sort.Slice(
			records,
			func(i, j int) bool {
				if params.DateDesc != nil && *params.DateDesc {
					return records[i].StartTime.After(records[j].StartTime)
				}

				return records[i].StartTime.Before(records[j].StartTime)
			},
		)
	}

	var page = params.Page
	if page <= 0 {
		page = 1
	}

	var (
		totalCount = len(records)
		start      = (page - 1) * params.Limit
		end        = start + params.Limit
	)
	if start > totalCount {
		start = totalCount
	}

	if end > totalCount {
		end = totalCount
	}

	if start < end {
		return &response.ListWithMapResp[[]*Item, string]{List: records[start:end], Count: int64(totalCount)}, nil
	}

	return &response.ListWithMapResp[[]*Item, string]{List: []*Item{}, Count: int64(totalCount)}, nil
}
