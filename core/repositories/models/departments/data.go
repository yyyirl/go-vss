package departments

import (
	"errors"
	"fmt"

	"skeyevss/core/pkg/functions"
)

type Item struct {
	*Departments

	ParentCascadeDepUniqueId string   `gorm:"column:-" json:"-"`
	RoleIds                  []uint64 `gorm:"column:-" json:"roleIds"`
	UseDBCache               bool     `json:"-"`
}

func NewItem() *Item {
	return new(Item)
}

func (i *Item) ConvToModel(call func(*Item) *Item) (*Departments, error) {
	if i.Departments == nil {
		return nil, nil
	}

	if call != nil {
		i = call(i)
	}

	if len(i.RoleIds) <= 0 {
		i.Departments.RoleIds = "[]"
	} else {
		roleIds, err := functions.ToString(i.RoleIds)
		if err != nil {
			return nil, err
		}

		i.Departments.RoleIds = roleIds
	}

	return i.Departments, nil
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

		if column == ColumnRoleIds {
			var roleIds []uint64
			if err := functions.ConvInterface(value, &roleIds); err != nil {
				return nil, fmt.Errorf("permission unique ids is invalid, input type: %T, needed []uint64", value)
			}

			b, err := functions.JSONMarshal(roleIds)
			if err != nil {
				return nil, err
			}

			input[column] = string(b)
		}

		if column == ColumnCascadeDepUniqueId {
			if value != nil {
				v, ok := value.(string)
				if ok && v == "" {
					input[column] = nil
				}
			}
		}
	}

	return input, nil
}
