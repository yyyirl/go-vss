package alarms

import (
	"errors"

	"github.com/mitchellh/mapstructure"

	"skeyevss/core/pkg/functions"
)

type Item struct {
	*Alarms

	UseDBCache bool `json:"-"`
}

func NewItem() *Item {
	return new(Item)
}

func (i *Item) ConvToModel(call func(*Item) *Item) (*Alarms, error) {
	if i.Alarms == nil {
		return nil, nil
	}

	if call != nil {
		i = call(i)
	}

	return i.Alarms, nil
}

func (i *Item) MapToModel(input map[string]interface{}) (*Item, error) {
	if input == nil {
		return nil, errors.New("input object is nil")
	}

	var model Alarms
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

	return &Item{Alarms: &model}, nil
}

func (*Item) CheckMap(input map[string]interface{}) (map[string]interface{}, error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}

	for column := range input {
		if !functions.Contains(column, Columns) {
			return nil, errors.New("column: " + column + " does not exist")
		}
	}

	return input, nil
}
