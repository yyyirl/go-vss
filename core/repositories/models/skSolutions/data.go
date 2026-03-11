package skSolutions

import (
	"errors"
	"skeyevss/core/pkg/orm"

	"skeyevss/core/pkg/functions"

	"github.com/mitchellh/mapstructure"
)

type Item struct {
	*SkSolutions

	UseDBCache bool `json:"-"`
}

func NewItem() *Item {
	return new(Item)
}

func (i *Item) ConvToModel(call func(*Item) *Item) (*SkSolutions, error) {
	if i.SkSolutions == nil {
		return nil, nil
	}

	if call != nil {
		i = call(i)
	}

	return i.SkSolutions, nil
}

func (i *Item) MapToModel(input map[string]interface{}) (*Item, error) {
	if input == nil {
		return nil, errors.New("input object is nil")
	}

	var model SkSolutions
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

	return &Item{SkSolutions: &model}, nil
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

func (*Item) ToUpdateMap(input []*orm.UpdateItem) (map[string]interface{}, error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}

	var maps = make(map[string]interface{})
	for _, item := range input {
		if !functions.Contains(item.Column, Columns) {
			return nil, errors.New("column: " + item.Column + " does not exist")
		}

		maps[item.Column] = item.Value
	}

	return maps, nil
}
