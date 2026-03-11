package devices

import (
	"errors"
	"fmt"

	"skeyevss/core/pkg/functions"
)

type Item struct {
	*Devices

	Sub            Subscription `json:"sub"`
	MSIds          []uint64     `json:"msIds"`
	DepIds         []uint64     `json:"depIds"`
	ChannelFilters []string     `json:"channelFilters"`
	UseDBCache     bool         `json:"-"`
}

func NewItem() *Item {
	return new(Item)
}

func (i *Item) ToMap() (map[string]interface{}, error) {
	if i.MSIds == nil {
		i.MSIds = []uint64{}
	}

	if i.ChannelFilters == nil {
		i.ChannelFilters = []string{}
	}

	if i.DepIds == nil {
		i.DepIds = []uint64{}
	}

	var record map[string]interface{}
	if err := functions.ConvInterface(i, &record); err != nil {
		return nil, err
	}

	return record, nil
}

func (i *Item) ConvToModel(call func(*Item) *Item) (*Devices, error) {
	if i.Devices == nil {
		return nil, nil
	}

	if call != nil {
		i = call(i)
	}

	if len(i.MSIds) <= 0 {
		i.Devices.MSIds = "[]"
	} else {
		val, err := functions.ToString(i.MSIds)
		if err != nil {
			return nil, err
		}

		i.Devices.MSIds = val
	}

	if len(i.ChannelFilters) <= 0 {
		i.Devices.ChannelFilters = "[]"
	} else {
		val, err := functions.ToString(i.ChannelFilters)
		if err != nil {
			return nil, err
		}

		i.Devices.ChannelFilters = val
	}

	if len(i.DepIds) <= 0 {
		i.Devices.DepIds = "[]"
	} else {
		depIds, err := functions.ToString(i.DepIds)
		if err != nil {
			return nil, err
		}

		i.Devices.DepIds = depIds
	}

	return i.Devices, nil
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

		if column == ColumnMSIds {
			var msIds []uint64
			if err := functions.ConvInterface(value, &msIds); err != nil {
				return nil, fmt.Errorf("device media server ids is invalid, input type: %T, needed []uint64", value)
			}

			b, err := functions.JSONMarshal(msIds)
			if err != nil {
				return nil, err
			}

			input[column] = string(b)
		}

		if column == ColumnDepIds {
			var depIds []uint64
			if err := functions.ConvInterface(value, &depIds); err != nil {
				return nil, fmt.Errorf("device department ids is invalid, input type: %T, needed []uint64", value)
			}

			b, err := functions.JSONMarshal(depIds)
			if err != nil {
				return nil, err
			}

			input[column] = string(b)
		}

		if column == ColumnChannelFilters {
			var channelFilters []string
			if err := functions.ConvInterface(value, &channelFilters); err != nil {
				return nil, fmt.Errorf("device channel filters is invalid, input type: %T, needed []string", value)
			}

			b, err := functions.JSONMarshal(channelFilters)
			if err != nil {
				return nil, err
			}

			input[column] = string(b)
		}
	}

	return input, nil
}

type TransportProtocol struct {
	Protocol          uint
	MediaProtocolMode uint // 0 udp 1 tcp
	MediaTransMode    string
	BitstreamIndex    uint
}

func (i *Item) TransportProtocol() *TransportProtocol {
	var data = &TransportProtocol{
		Protocol:          0,
		MediaProtocolMode: 0,
		MediaTransMode:    "passive",
		BitstreamIndex:    i.BitstreamIndex,
	}

	switch i.MediaTransMode {
	case MediaTransMode_1:
		data.MediaTransMode = "passive"
		data.MediaProtocolMode = 1

	case MediaTransMode_2:
		data.MediaTransMode = "active"
		data.MediaProtocolMode = 1

	case MediaTransMode_0:
		data.MediaTransMode = "passive"
		data.Protocol = 1
	}

	return data
}

type CrontabItem struct {
	StreamName string `json:"streamName"`
	RetryCount uint   `json:"retryCount"`
}

func (c CrontabItem) ToString() string {
	data, _ := functions.JSONMarshal(c)
	return string(data)
}
