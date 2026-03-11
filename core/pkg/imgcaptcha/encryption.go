/**
 * @Author:         yi
 * @Description:    encryption
 * @Version:        1.0.0
 * @Date:           2023/9/17 11:36
 */
package imgcaptcha

import (
	"errors"

	"skeyevss/core/pkg/functions"
)

func DotsEncryption(key string, expire int64, dots Dots) (string, error) {
	b, err := functions.JSONMarshal(&Transfer{
		Dots:   dots,
		Expire: expire,
	})
	if err != nil {
		return "", err
	}

	return functions.NewCrypto([]byte(key)).Encrypt(b)
}

func DotsDecryption(secret, key, dots string) error {
	encrypt, err := functions.NewCrypto([]byte(secret)).Decrypt(key)
	if err != nil {
		return err
	}

	var cpt Transfer
	if err := functions.JSONUnmarshal([]byte(encrypt), &cpt); err != nil {
		return err
	}

	if cpt.Expire-functions.NewTimer().Now() < 0 {
		// return errors.New("验证码已过期")
	}

	var ipt []Coordinates
	if err := functions.JSONUnmarshal([]byte(dots), &ipt); err != nil {
		return err
	}

	if len(ipt) != len(cpt.Dots) {
		return errors.New("非法请求")
	}

	var correctDots []Coordinates
	for _, item := range cpt.Dots {
		correctDots = append(
			correctDots,
			Coordinates{
				MinX: item.Dx,
				MinY: item.Dy - item.Height,
				X:    item.Dx + item.Width,
				Y:    item.Dy + item.Height,
			},
		)
	}

	for key, item := range correctDots {
		if ipt[key].X < item.MinX || ipt[key].X > item.X {
			return errors.New("坐标X不匹配")
		}

		if ipt[key].Y < item.MinY || ipt[key].Y > item.Y {
			return errors.New("坐标Y不匹配")
		}
	}

	return nil
}
