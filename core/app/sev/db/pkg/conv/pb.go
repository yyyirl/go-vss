package conv

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/pkg/orm"
)

type pbParams struct {
	mode string // dev prod
}

func (t *pbParams) slice(input []interface{}) ([]*anypb.Any, error) {
	var records []*anypb.Any
	for _, item := range input {
		v, err := t.interfaceToPBAny(item)
		if err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		records = append(records, v)
	}

	return records, nil
}

func (t *pbParams) interfaceToPBAny(value interface{}) (*anypb.Any, error) {
	if value == nil {
		return nil, makeErr(t.mode, "nil value cannot be converted to Any")
	}

	if anyVal, ok := value.(*anypb.Any); ok {
		return anyVal, nil
	}

	if msg, ok := value.(proto.Message); ok {
		v, err := anypb.New(msg)
		if err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return v, nil
	}

	switch v := value.(type) {
	case string:
		d, err := anypb.New(wrapperspb.String(v))
		if err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return d, nil

	case int, int32, int64, int8, int16:
		d, err := anypb.New(wrapperspb.Int64(reflect.ValueOf(v).Int()))
		if err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return d, nil

	case uint, uint32, uint64, uint8, uint16:
		d, err := anypb.New(wrapperspb.UInt64(reflect.ValueOf(v).Uint()))
		if err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return d, nil

	case float32, float64:
		d, err := anypb.New(wrapperspb.Double(reflect.ValueOf(v).Float()))
		if err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return d, nil

	case bool:
		d, err := anypb.New(wrapperspb.Bool(v))
		if err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return d, nil

	case []byte:
		d, err := anypb.New(wrapperspb.Bytes(v))
		if err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return d, nil
	}

	// map[string]interface{}（转换为 structpb.Struct）
	if reflect.TypeOf(value).Kind() == reflect.Map {
		if m, ok := value.(map[string]interface{}); ok {
			s, err := structpb.NewStruct(m)
			if err != nil {
				return nil, fmt.Errorf("failed to convert map to Struct: %v", err)
			}
			d, err := anypb.New(s)
			if err != nil {
				return nil, makeErr(t.mode, err.Error())
			}

			return d, nil
		}
	}

	// 切片/数组（转换为 structpb.ListValue）
	if reflect.TypeOf(value).Kind() == reflect.Slice || reflect.TypeOf(value).Kind() == reflect.Array {
		var (
			slice = reflect.ValueOf(value)
			list  = make([]interface{}, slice.Len())
		)
		for i := 0; i < slice.Len(); i++ {
			list[i] = slice.Index(i).Interface()
		}

		lv, err := structpb.NewList(list)
		if err != nil {
			return nil, makeErr(t.mode, err.Error())
		}
		d, err := anypb.New(lv)
		if err != nil {
			return nil, makeErr(t.mode, err.Error())
		}

		return d, nil
	}

	return nil, makeErr(t.mode, fmt.Sprintf("unsupported type: %T", value))
}

func (t *pbParams) conditions(input []*orm.ConditionItem) ([]*db.XConditionItem, error) {
	var conditions []*db.XConditionItem
	if len(input) <= 0 {
		return conditions, nil
	}
	for _, item := range input {
		var (
			err    error
			record = &db.XConditionItem{
				Column:          item.Column,
				Operator:        item.Operator,
				UseNil:          item.UseNil,
				LogicalOperator: item.LogicalOperator,
				Columns:         item.Columns,
			}
		)

		if item.Value != nil {
			record.Value, err = t.interfaceToPBAny(item.Value)
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

			record.Original = &db.XConditionOriginalItem{
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
