package videoProjects

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"

	"skeyevss/core/pkg/functions"
)

type Item struct {
	*VideoProjects

	ChannelUniqueIds []uint64 `gorm:"column:-" json:"channelUniqueIds"`
	UseDBCache       bool     `json:"-"`
}

func NewItem() *Item {
	return new(Item)
}

func (i *Item) ConvToModel(call func(*Item) *Item) (*VideoProjects, error) {
	if i.VideoProjects == nil {
		return nil, nil
	}

	channelUniqueIds, err := functions.JSONMarshal(i.ChannelUniqueIds)
	if err != nil {
		return nil, err
	}

	i.VideoProjects.ChannelUniqueIds = string(channelUniqueIds)
	if len(i.VideoProjects.ChannelUniqueIds) <= 0 {
		i.VideoProjects.ChannelUniqueIds = "[]"
	}

	if call != nil {
		i = call(i)
	}

	return i.VideoProjects, nil
}

func (i *Item) MapToModel(input map[string]interface{}) (*Item, error) {
	if input == nil {
		return nil, errors.New("input object is nil")
	}

	var model VideoProjects
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			DecodeHook: mapstructure.DecodeHookFunc(functions.MapStructureHook),
			Result:     &model,
			// TagName:    "mapstructure",
		},
	)
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(input); err != nil {
		return nil, err
	}

	return &Item{VideoProjects: &model}, nil
}

func (*Item) CheckMap(input map[string]interface{}) (map[string]interface{}, error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}

	for column, value := range input {
		if !functions.Contains(column, Columns) {
			return nil, errors.New("column: " + column + " does not exist")
		}

		if column == ColumnChannelUniqueIds {
			var uniqueIds []uint64
			if err := functions.ConvInterface(value, &uniqueIds); err != nil {
				return nil, fmt.Errorf("channel unique ids is invalid, input type: %T, needed []uint64", value)
			}

			b, err := functions.JSONMarshal(uniqueIds)
			if err != nil {
				return nil, err
			}

			input[column] = string(b)
		}
	}

	return input, nil
}
