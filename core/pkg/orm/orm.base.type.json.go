package orm

import (
	"bytes"
	"context"
	"errors"
	"reflect"

	"gorm.io/gorm/schema"
)

type DBJson []byte

func (j DBJson) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	if j.IsNull() {
		return nil, nil
	}

	return string(j), nil
}

func (j DBJson) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	if dbValue == nil {
		j = nil
		return nil
	}

	s, ok := dbValue.([]byte)
	if !ok {
		return errors.New("invalid Scan Source")
	}

	j = append(j[0:0], s...)

	return nil
}

func (m DBJson) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}

	return m, nil
}

func (m *DBJson) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("null point exception")
	}

	*m = append((*m)[0:0], data...)

	return nil
}

func (j DBJson) IsNull() bool {
	return len(j) == 0 || string(j) == "null"
}

func (j DBJson) Equals(j1 DBJson) bool {
	return bytes.Equal(j, j1)
}

// 注册解析器
//	schema.RegisterSerializer("x-json", orm.DBJson{})
