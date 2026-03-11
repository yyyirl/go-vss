// @Title        main
// @Description  main
// @Create       yiyiyi 2025/8/11 14:23

package onvif

import (
	"fmt"

	goonvif "github.com/use-go/onvif"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/repositories/models/devices"
)

func GetDevWithDevice(item *devices.Item) (*goonvif.Device, error) {
	addrRes, err := functions.ExtractBaseURL(item.Address)
	if err != nil {
		return nil, err
	}

	dev, err := goonvif.NewDevice(goonvif.DeviceParams{
		Xaddr:    fmt.Sprintf("%s:%d", addrRes.IP, addrRes.Port),
		Username: item.Username,
		Password: item.Password,
	})
	if err != nil {
		return nil, err
	}

	return dev, nil
}

type GetDevParams struct {
	Username string
	Password string
	IP       string
	Port     uint
}

func GetDev(data *GetDevParams) (*goonvif.Device, error) {
	dev, err := goonvif.NewDevice(goonvif.DeviceParams{
		Xaddr:    fmt.Sprintf("%s:%d", data.IP, data.Port),
		Username: data.Username,
		Password: data.Password,
	})
	if err != nil {
		return nil, err
	}

	return dev, nil
}
