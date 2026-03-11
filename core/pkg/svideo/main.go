// @Title        main
// @Description  main
// @Create       yiyiyi 2025/8/18 11:46

package svideo

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"skeyevss/core/pkg/functions"
)

type (
	SVideo struct {
		// 文件存储目录 // SKEYEVSS_SAVE_VIDEO_DIR
		SaveDir string
	}
	DateRange struct {
		Start,
		End uint64
	}

	Params struct {
		DeviceUniqueId,
		ChannelUniqueId string
		Page  int
		Limit int
		Date  *DateRange

		DateDesc,
		UrgentDesc,
		Urgent *bool
	}

	Item struct {
		ID              uint64 `json:"id"`
		DeviceUniqueId  string `json:"deviceUniqueId"`
		ChannelUniqueId string `json:"channelUniqueId"`
		Filename        string `json:"filename"`
		Start           int64  `json:"start"`
		End             int64  `json:"end"`
		Address         string `json:"address"`
		Snapshot        string `json:"snapshot"`
		Urgent          bool   `json:"urgent"`

		StartTime time.Time `json:"-"`
		EndTime   time.Time `json:"-"`
	}
)

func NewSVideo(saveDir string) *SVideo {
	return &SVideo{SaveDir: saveDir}
}

func (s *SVideo) MakeFileName(data *Item) string {
	var (
		timer = functions.NewTimer()
		start = timer.FormatTimestamp(data.Start, functions.TimeFormatYmdhis)
		end   = timer.FormatTimestamp(data.End, functions.TimeFormatYmdhis)
	)
	if data.Urgent {
		return fmt.Sprintf("%s|%s|%s|%s|1.mp4", data.DeviceUniqueId, data.ChannelUniqueId, start, end)
	}

	return fmt.Sprintf("%s|%s|%s|%s.mp4", data.DeviceUniqueId, data.ChannelUniqueId, start, end)
}

func (s *SVideo) Parse(filename string, index uint64) *Item {
	var (
		base   = strings.TrimSuffix(filepath.Base(filename), ".mp4")
		parts  = strings.Split(base, "|")
		length = len(parts)
	)
	if length < 4 {
		return nil
	}

	// 解析时间
	startTime, err := time.Parse("2006-01-02-15-04-05", parts[2])
	if err != nil {
		return nil
	}

	endTime, err := time.Parse("2006-01-02-15-04-05", parts[3])
	if err != nil {
		return nil
	}

	var urgent = false
	if length >= 5 {
		urgent = parts[4] != "0"
	}

	return &Item{
		ID:              index,
		DeviceUniqueId:  parts[0],
		ChannelUniqueId: parts[1],
		Filename:        filename,
		Start:           startTime.UnixMilli(),
		End:             endTime.UnixMilli(),
		Snapshot:        strings.Replace(filename, ".mp4", ".jpg", -1),
		Urgent:          urgent,

		StartTime: startTime,
		EndTime:   endTime,
	}
}
