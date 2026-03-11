package settings

import (
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"

	"skeyevss/core/constants"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/tps"
)

type Item struct {
	*Settings

	Content *Content `json:"content"`

	UseDBCache bool `json:"-"`
}

func NewItem() *Item {
	return new(Item)
}

func (i *Item) ConvToModel(call func(*Item) *Item) (*Settings, error) {
	if i.Settings == nil {
		return nil, nil
	}

	if call != nil {
		i = call(i)
	}

	if i.Content == nil {
		i.Settings.Content = "{}"
	} else {
		units, err := functions.ToString(i.Content)
		if err != nil {
			return nil, err
		}

		i.Settings.Content = units
	}

	return i.Settings, nil
}

func (i *Item) MapToModel(input map[string]interface{}) (*Item, error) {
	if input == nil {
		return nil, errors.New("input object is nil")
	}

	var model Settings
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

	return &Item{Settings: &model}, nil
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

type ItemCorrectionParams struct {
	BaseConf tps.YamlSevBaseConfig
	SipConf  tps.YamlSip

	InternalIp,
	ExternalIp string
}

func (i *Item) ItemCorrection(_ *ItemCorrectionParams) {
	if i.Content.MediaServerVideoPlayAddressType == "" {
		i.Content.MediaServerVideoPlayAddressType = constants.VideoPlayAddressTypeWsFlv
	}

	if i.Content.BanIp != "" {
		var list []string
		for _, item := range strings.Split(i.Content.BanIp, "\n") {
			list = append(list, strings.Split(strings.TrimSpace(item), " ")...)
		}

		var records []string
		for _, item := range list {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}

			records = append(records, item)
		}

		i.Content.BanIp = strings.Join(records, "\n")
	}
}
