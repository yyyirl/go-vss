package common

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"

	"skeyevss/core/localization"
	"skeyevss/core/pkg/elasticsearch"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/tps"
)

type Conv struct{}

func NewConv() *Conv {
	return new(Conv)
}

func (*Conv) ReqParamsToOrmFindParams(data interface{}) (*orm.ReqParams, *response.HttpErr) {
	if data == nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Str("请求参数不能为空"), localization.M0001)
	}

	var req orm.ReqParams
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			TagName: "ms",
			Result:  &req,
		},
	)
	if err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0009)
	}

	if err := decoder.Decode(data); err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001)
	}

	// 后台网页特殊处理
	if len(req.Data) > 0 {
		req.DataRecord = make(map[string]interface{})
		for _, item := range req.Data {
			req.DataRecord[item.Column] = item.Value
		}
	}

	return &req, nil
}

func (*Conv) ReqBytesParamsToOrmFindParams(data []byte) (*orm.ReqParams, error) {
	if data == nil {
		return nil, errors.New("请求参数不能为空")
	}

	var req orm.ReqParams
	if err := functions.JSONUnmarshal(data, &req); err != nil {
		return nil, err
	}

	// 后台网页特殊处理
	if len(req.Data) > 0 {
		req.DataRecord = make(map[string]interface{})
		for _, item := range req.Data {
			req.DataRecord[item.Column] = item.Value
		}
	}

	return &req, nil
}

func (*Conv) ReqParamsToEsFindParams(data interface{}) (*elasticsearch.ReqParams, *response.HttpErr) {
	if data == nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Str("请求参数不能为空"), localization.M0001)
	}

	var req elasticsearch.ReqParams
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			TagName: "ms",
			Result:  &req,
		},
	)
	if err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0009)
	}

	if err := decoder.Decode(data); err != nil {
		return nil, response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001)
	}

	// 后台网页特殊处理
	if len(req.Data) > 0 {
		req.DataRecord = make(map[string]interface{})
		for _, item := range req.Data {
			req.DataRecord[item.Column] = item.Value
		}
	}

	return &req, nil
}

func (*Conv) MapToStruct(data, value interface{}) *response.HttpErr {
	if data == nil {
		return response.MakeError(response.NewHttpRespMessage().Str("请求参数不能为空"), localization.M0001)
	}

	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			TagName: "ms",
			Result:  &value,
		},
	)
	if err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0009)
	}

	if err := decoder.Decode(data); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0001)
	}

	return nil
}

// records, err := common.NewConv().ReqParamsParseRecords(req.Records, reflect.TypeOf(new(tps.VectorCharInputItem)))
func (*Conv) ReqParamsParseRecords(input []interface{}, outputType reflect.Type) (interface{}, *response.HttpErr) {
	var records = reflect.MakeSlice(reflect.SliceOf(outputType), len(input), len(input))
	for i, v := range input {
		var value = reflect.ValueOf(v)
		if value.Type().ConvertibleTo(outputType) {
			records.Index(i).Set(value.Convert(outputType))
		} else {
			return nil, response.MakeError(response.NewHttpRespMessage().Err(fmt.Errorf("element %d cannot be converted to %s", i, outputType.String())), localization.M0001)
		}
	}

	return records.Interface(), nil
}

func (*Conv) LangSliceToMap(data []*tps.Lang) map[uint]*tps.Lang {
	var maps = make(map[uint]*tps.Lang, len(data))
	for _, item := range data {
		maps[item.ID] = item
	}

	return maps
}
