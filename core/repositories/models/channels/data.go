package channels

import (
	"errors"
	"fmt"

	"skeyevss/core/pkg/functions"
)

type (
	RangeDate struct {
		Start uint64 `json:"start"`
		End   uint64 `json:"end"`
	}

	VideoItem struct {
		Date *RangeDate `json:"date"`
		Path string     `json:"path"`
	}
)

type Item struct {
	*Channels
	UseDBCache bool `json:"-"`

	Videos      []*VideoItem           `gorm:"column:-" json:"videos"`
	Screenshots []string               `gorm:"column:-" json:"screenshots"`
	Original    map[string]interface{} `gorm:"column:-" json:"original"`
	DepIds      []uint64               `gorm:"column:-" json:"depIds"`
}

func NewItem() *Item {
	return new(Item)
}

func (i *Item) ConvToModel(call func(*Item) *Item) (*Channels, error) {
	if i.Channels == nil {
		return nil, nil
	}

	if call != nil {
		i = call(i)
	}

	if i.Original == nil {
		i.Channels.Original = "{}"
	} else {
		original, err := functions.ToString(i.Original)
		if err != nil {
			return nil, err
		}

		i.Channels.Original = original
	}

	if len(i.Videos) <= 0 {
		i.Channels.Videos = "[]"
	} else {
		videos, err := functions.ToString(i.Videos)
		if err != nil {
			return nil, err
		}

		i.Channels.Videos = videos
	}

	if len(i.Screenshots) <= 0 {
		i.Channels.Screenshots = "[]"
	} else {
		screenshots, err := functions.ToString(i.Screenshots)
		if err != nil {
			return nil, err
		}

		i.Channels.Screenshots = screenshots
	}

	if len(i.DepIds) <= 0 {
		i.Channels.DepIds = "[]"
	} else {
		depIds, err := functions.ToString(i.DepIds)
		if err != nil {
			return nil, err
		}

		i.Channels.DepIds = depIds
	}

	return i.Channels, nil
}

func (i *Item) MapToModel(input map[string]interface{}) (*Item, error) {
	if input == nil {
		return nil, errors.New("input object is nil")
	}

	b, err := functions.JSONMarshal(input)
	if err != nil {
		return nil, err
	}

	var model Item
	if err := functions.JSONUnmarshal(b, &model); err != nil {
		return nil, err
	}

	return &model, nil
}

func (*Item) CheckMap(input map[string]interface{}) (map[string]interface{}, error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}

	for column, value := range input {
		if !functions.Contains(column, Columns) {
			return nil, errors.New("column: " + column + " does not exist")
		}

		if column == ColumnOriginal {
			var original map[string]interface{}
			if err := functions.ConvInterface(value, &original); err != nil {
				return nil, fmt.Errorf("channels original is invalid, input type: %T, needed map[string]interface", value)
			}

			b, err := functions.JSONMarshal(original)
			if err != nil {
				return nil, err
			}

			input[column] = string(b)
		}

		if column == ColumnVideos {
			var videos []*VideoItem
			if err := functions.ConvInterface(value, &videos); err != nil {
				return nil, fmt.Errorf("channel videos is invalid, input type: %T, needed []*VideoItem, err: %s", value, err)
			}

			if len(videos) <= 0 {
				videos = []*VideoItem{}
			}

			b, err := functions.JSONMarshal(videos)
			if err != nil {
				return nil, err
			}

			input[column] = string(b)
		}

		if column == ColumnScreenshots {
			var screenshots []string
			if err := functions.ConvInterface(value, &screenshots); err != nil {
				return nil, fmt.Errorf("channel screenshots is invalid, input type: %T, needed []string, err: %s", value, err)
			}

			if len(screenshots) <= 0 {
				screenshots = []string{}
			}

			b, err := functions.JSONMarshal(screenshots)
			if err != nil {
				return nil, err
			}

			input[column] = string(b)
		}

		if column == ColumnDepIds {
			var depIds []uint64
			if err := functions.ConvInterface(value, &depIds); err != nil {
				return nil, fmt.Errorf("channel depIds is invalid, input type: %T, needed []uint64, err: %s", value, err)
			}

			if len(depIds) <= 0 {
				depIds = []uint64{}
			}

			b, err := functions.JSONMarshal(depIds)
			if err != nil {
				return nil, err
			}

			input[column] = string(b)
		}
	}

	return input, nil
}
