package functions

import (
	"reflect"

	"skeyevss/core/tps"
)

func Contains[T comparable](x T, arr []T) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}

	return false
}

func ConvNumbers[T, V tps.Number](data []T) (final []V) {
	var list []V
	for _, v := range data {
		list = append(list, V(v))
	}

	return list
}

func ArrUnique[T comparable](data []T) (final []T) {
	if len(data) <= 0 {
		return nil
	}

	var (
		result = make([]T, 0, len(data))
		temp   = make(map[T]struct{}, 0)
	)
	for _, v := range data {
		if _, ok := temp[v]; ok {
			continue
		}

		temp[v] = struct{}{}
		result = append(result, v)
	}

	return result
}

func ArrUniqueWithCall[T any](data []T, getUniqueIdKey func(item T) string) (final []T) {
	if len(data) <= 0 {
		return nil
	}

	var (
		result = make([]T, 0, len(data))
		temp   = make(map[string]struct{}, 0)
	)
	for _, v := range data {
		var key = getUniqueIdKey(v)
		if _, ok := temp[key]; ok {
			continue
		}

		temp[key] = struct{}{}
		result = append(result, v)
	}

	return result
}

func MapArrUniqueWithCall[K comparable, T any](data map[K][]T, getUniqueIdKey func(item T) string) map[K][]T {
	if len(data) <= 0 {
		return nil
	}

	for k, v := range data {
		data[k] = ArrUniqueWithCall(v, getUniqueIdKey)
	}

	return data
}

func MapToMapInterface[K comparable, T any](data map[K]T) map[K]interface{} {
	var m = make(map[K]interface{}, len(data))
	for k, v := range data {
		m[k] = v
	}

	return m
}

func ArrFilter[T any](data []T, call func(item T) bool) (final []T) {
	if len(data) <= 0 {
		return nil
	}

	var result = make([]T, 0, len(data))
	for _, v := range data {
		if !call(v) {
			continue
		}

		result = append(result, v)
	}

	return result
}

func SliceToSliceAny[T tps.SliceCompareValue](data []T) []any {
	var (
		list = make([]any, len(data))
		i    = 0
	)
	for _, item := range data {
		list[i] = item
		i++
	}

	return list
}

func Range[T any](data []T, call func(item T) T) []T {
	for k, item := range data {
		data[k] = call(item)
	}

	return data
}

// 检测数组中是否有重复元素
func HasDuplicates[T tps.Simple](arr []T) bool {
	var seen = make(map[T]struct{})
	for _, item := range arr {
		if _, exists := seen[item]; exists {
			return true
		}
		seen[item] = struct{}{}
	}

	return false
}

func IsNumberSlice(s interface{}) bool {
	var val = reflect.ValueOf(s)
	if val.Kind() != reflect.Slice {
		return false
	}

	var (
		elemType = val.Type().Elem()
		kind     = elemType.Kind()
	)

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true

	default:
		return false
	}
}

func IsStringSlice(s interface{}) bool {
	var val = reflect.ValueOf(s)
	if val.Kind() != reflect.Slice {
		return false
	}

	var (
		elemType = val.Type().Elem()
		kind     = elemType.Kind()
	)

	switch kind {
	case reflect.String:
		return true

	default:
		return false
	}
}
