package cascade

import (
	"errors"
	"fmt"

	"skeyevss/core/pkg/functions"
)

type RegisterState uint

type RelationItem struct {
	Parental bool   `json:"parental"`
	UniqueId string `json:"uniqueId"`
}

type Item struct {
	*Cascade

	Relations  []RelationItem `gorm:"column:-" json:"relations"`
	UseDBCache bool           `json:"-"`

	// sip state
	XSipRegisterState RegisterState `gorm:"column:-" json:"-"`
	XSipAuth          string        `gorm:"column:-" json:"-"`
}

func NewItem() *Item {
	return new(Item)
}

func (i *Item) ConvToModel(call func(*Item) *Item) (*Cascade, error) {
	if i.Cascade == nil {
		return nil, nil
	}

	if call != nil {
		i = call(i)
	}

	if len(i.Relations) <= 0 {
		i.Cascade.Relations = "[]"
	} else {
		relations, err := functions.ToString(i.Relations)
		if err != nil {
			return nil, err
		}

		i.Cascade.Relations = relations
	}

	return i.Cascade, nil
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

		if column == ColumnRelations {
			var relations []RelationItem
			if err := functions.ConvInterface(value, &relations); err != nil {
				return nil, fmt.Errorf("Cascade relations is invalid, input type: %T, needed []RelationItem, err: %s", value, err)
			}

			if len(relations) <= 0 {
				relations = []RelationItem{}
			}

			b, err := functions.JSONMarshal(relations)
			if err != nil {
				return nil, err
			}

			input[column] = string(b)
		}
	}

	return input, nil
}

func (i *Item) ProtocolToString() string {
	return ProtocolMaps[i.Protocol]
}

func (i *Item) CommandTransportToString() string {
	return ProtocolMaps[i.CommandTransport]
}

func (i *Item) MakeGBCRegisterExecutingKey() string {
	return fmt.Sprintf("%d-%s-%d", i.ID, i.SipIp, i.State)
}

func (i *Item) DelayRegisterTimeout() uint {
	if i.RegisterTimeout <= 30 {
		i.RegisterTimeout = 30
	}

	return i.RegisterTimeout - 5
}
