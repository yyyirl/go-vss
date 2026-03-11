package crontab

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"skeyevss/core/pkg/functions"
)

type Item struct {
	*Crontab

	UseDBCache bool     `json:"-"`
	Logs       []string `json:"logs"`
}

func NewItem() *Item {
	return new(Item)
}

func (i *Item) ConvToModel(call func(*Item) *Item) (*Crontab, error) {
	if i.Crontab == nil {
		return nil, nil
	}

	if len(i.Logs) <= 0 {
		i.Crontab.Logs = "[]"
	} else {
		data, err := functions.ToString(i.Logs)
		if err != nil {
			return nil, err
		}

		i.Crontab.Logs = data
	}

	if call != nil {
		i = call(i)
	}

	return i.Crontab, nil
}

func (i *Item) MapToModel(input map[string]interface{}) (*Item, error) {
	if input == nil {
		return nil, errors.New("input object is nil")
	}

	var model Crontab
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

	return &Item{Crontab: &model}, nil
}

func (*Item) CheckMap(input map[string]interface{}) (map[string]interface{}, error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}

	for column, value := range input {
		if !functions.Contains(column, Columns) {
			return nil, errors.New("column: " + column + " does not exist")
		}

		if column == ColumnLogs {
			var data []string
			if err := functions.ConvInterface(value, &data); err != nil {
				return nil, fmt.Errorf("crontab logs is invalid, input type: %T, needed []string", value)
			}

			b, err := functions.JSONMarshal(data)
			if err != nil {
				return nil, err
			}

			input[column] = string(b)
		}
	}

	return input, nil
}
