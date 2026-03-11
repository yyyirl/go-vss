package roles

import (
	"errors"
	"fmt"

	"skeyevss/core/pkg/functions"
)

type Item struct {
	*Roles

	PermissionUniqueIds []string `gorm:"column:-" json:"permissionUniqueIds"`
	UseDBCache          bool     `json:"-"`
}

func NewItem() *Item {
	return new(Item)
}

func (i *Item) ConvToModel(call func(*Item) *Item) (*Roles, error) {
	if i.Roles == nil {
		return nil, nil
	}

	permissionUniqueIds, err := functions.JSONMarshal(i.PermissionUniqueIds)
	if err != nil {
		return nil, err
	}

	i.Roles.PermissionUniqueIds = string(permissionUniqueIds)
	if len(i.Roles.PermissionUniqueIds) <= 0 {
		i.Roles.PermissionUniqueIds = "[]"
	}

	if call != nil {
		i = call(i)
	}

	return i.Roles, nil
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

		if column == ColumnPermissionUniqueIds {
			var uniqueIds []string
			if err := functions.ConvInterface(value, &uniqueIds); err != nil {
				return nil, fmt.Errorf("permission unique ids is invalid, input type: %T, needed []string", value)
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
