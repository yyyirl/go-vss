package conv

import (
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/pkg/orm"
)

type ormParams struct {
	mode string // dev prod
}

func (t *ormParams) slice(input []*anypb.Any) ([]interface{}, error) {
	var records []interface{}
	for _, item := range input {
		v, err := t.pbAnyToInterface(item)
		if err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		records = append(records, v)
	}

	return records, nil
}

func (t *ormParams) pbAnyToInterface(data *anypb.Any) (interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("nil Any value")
	}

	switch data.TypeUrl {
	case "type.googleapis.com/google.protobuf.StringValue":
		var s wrapperspb.StringValue
		if err := data.UnmarshalTo(&s); err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return s.Value, nil

	case "type.googleapis.com/google.protobuf.Int64Value":
		var i wrapperspb.Int64Value
		if err := data.UnmarshalTo(&i); err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return i.Value, nil

	case "type.googleapis.com/google.protobuf.UInt64Value":
		var u wrapperspb.UInt64Value
		if err := data.UnmarshalTo(&u); err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return u.Value, nil

	case "type.googleapis.com/google.protobuf.DoubleValue":
		var d wrapperspb.DoubleValue
		if err := data.UnmarshalTo(&d); err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return d.Value, nil

	case "type.googleapis.com/google.protobuf.BoolValue":
		var b wrapperspb.BoolValue
		if err := data.UnmarshalTo(&b); err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return b.Value, nil

	case "type.googleapis.com/google.protobuf.BytesValue":
		var b wrapperspb.BytesValue
		if err := data.UnmarshalTo(&b); err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return b.Value, nil

	case "type.googleapis.com/google.protobuf.Struct":
		var s structpb.Struct
		if err := data.UnmarshalTo(&s); err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return s.AsMap(), nil

	case "type.googleapis.com/google.protobuf.ListValue":
		var l structpb.ListValue
		if err := data.UnmarshalTo(&l); err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return l.AsSlice(), nil

	default:
		msg, err := data.UnmarshalNew()
		if err != nil {
			return nil, makeErr(t.mode, fmt.Sprintf("failed to unmarshal message from Any: %v", err))
		}

		return msg, nil
	}
}

func (t *ormParams) conditions(input []*db.XConditionItem) ([]*orm.ConditionItem, error) {
	var conditions []*orm.ConditionItem
	if len(input) <= 0 {
		return conditions, nil
	}
	for _, item := range input {
		var (
			err    error
			record = &orm.ConditionItem{
				Column:          item.Column,
				Operator:        item.Operator,
				UseNil:          item.UseNil,
				LogicalOperator: item.LogicalOperator,
				Columns:         item.Columns,
			}
		)

		if item.Value != nil {
			record.Value, err = t.pbAnyToInterface(item.Value)
			if err != nil {
				return nil, makeErr(t.mode, err.Error())
			}
		}

		if len(item.Values) > 0 {
			record.Values, err = t.slice(item.Values)
			if err != nil {
				return nil, makeErr(t.mode, err.Error())
			}
		}

		if item.Original != nil {
			if len(item.Original.Values) <= 0 {
				return nil, makeErr(t.mode, "no values found")
			}

			values, err := t.slice(item.Original.Values)
			if err != nil {
				return nil, makeErr(t.mode, err.Error())
			}

			record.Original = &orm.ConditionOriginalItem{
				Query:  item.Original.Query,
				Values: values,
			}
		}

		if len(item.Inner) > 0 {
			record.Inner, err = t.conditions(item.Inner)
			if err != nil {
				return nil, makeErr(t.mode, err.Error())
			}
		}

		conditions = append(conditions, record)
	}

	return conditions, nil
}
