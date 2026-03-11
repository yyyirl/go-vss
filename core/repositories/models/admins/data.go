package admins

import (
	"errors"
	"fmt"

	"skeyevss/core/pkg/functions"
)

type Item struct {
	*Admins

	UseDBCache bool     `json:"-"`
	DepIds     []uint64 `gorm:"column:-" json:"depIds"`
}

func NewItem() *Item {
	return new(Item)
}

func (i *Item) ConvToModel(call func(*Item) *Item) (*Admins, error) {
	if i.Admins == nil {
		return nil, errors.New("item is nil")
	}

	if call != nil {
		i = call(i)
	}

	if len(i.DepIds) <= 0 {
		i.Admins.DepIds = "[]"
	} else {
		depIds, err := functions.ToString(i.DepIds)
		if err != nil {
			return nil, err
		}

		i.Admins.DepIds = depIds
	}

	return i.Admins, nil
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

		if column == ColumnPassword {
			if v, ok := value.(string); ok {
				input[column], _ = functions.GeneratePwd(v)
			} else {
				return nil, errors.New("column: " + column + " is not a string")
			}
		}

		if column == ColumnDepIds {
			var depIds []uint64
			if err := functions.ConvInterface(value, &depIds); err != nil {
				return nil, fmt.Errorf("permission unique ids is invalid, input type: %T, needed []uint64", value)
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

// func (a *Admins) AfterFind(db *gorm.DB) (err error) {
// 	var querySql = db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
// 	fmt.Printf("\n sql: %+v \n", querySql)
// 	fmt.Printf("\n un: %+v \n", db.Statement.Context.Value("uniqueId"))
//
// 	return
// }
