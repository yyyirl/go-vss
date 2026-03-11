package dictionaries

import (
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"

	"skeyevss/core/pkg/functions"
)

type Item struct {
	*Dictionaries

	UseDBCache bool `json:"-"`
}

func NewItem() *Item {
	return new(Item)
}

func (i *Item) ConvToModel(call func(*Item) *Item) (*Dictionaries, error) {
	if i.Dictionaries == nil {
		return nil, nil
	}

	if call != nil {
		i = call(i)
	}

	return i.Dictionaries, nil
}

func (i *Item) MapToModel(input map[string]interface{}) (*Item, error) {
	if input == nil {
		return nil, errors.New("input object is nil")
	}

	var model Dictionaries
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

	return &Item{Dictionaries: &model}, nil
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

func (i *Item) GetMultiValue() []string {
	var list []string
	for _, item := range strings.Split(i.Dictionaries.MultiValue, "\n") {
		list = append(list, strings.TrimSpace(item))
	}

	return list
}

// func (*Item) ParseCategoryTrees(data []*categories.Item[int]) (map[string]*categories.Item[int], error) {
// 	var (
// 		maps = make(map[string]*categories.Item[int])
// 		call func(list []*categories.Item[int]) ([]*categories.Item[int], error)
// 	)
//
// 	call = func(list []*categories.Item[int]) ([]*categories.Item[int], error) {
// 		for _, item := range list {
// 			raw, ok := item.Raw.(map[string]interface{})
// 			if !ok {
// 				return nil, errors.New("字典原始数据解析失败")
// 			}
//
// 			v, err := NewItem().MapToModel(raw)
// 			if err != nil {
// 				return nil, errors.New("字典原始数据解析失败[1]")
// 			}
//
// 			item.Raw = v
// 			if len(item.Children) > 0 {
// 				item.Children, err = call(item.Children)
// 				if err != nil {
// 					return nil, errors.New("字典原始数据解析失败[3]")
// 				}
// 			}
//
// 			if item.Pid == 0 {
// 				maps[v.UniqueId] = item
// 			}
// 		}
//
// 		return list, nil
// 	}
//
// 	if _, err := call(data); err != nil {
// 		return nil, err
// 	}
//
// 	return maps, nil
// }
