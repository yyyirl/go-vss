/**
 * @Author:         yi
 * @Description:    main
 * @Version:        1.0.0
 * @Date:           2023/9/15 16:40
 */
package imgcaptcha

import (
	"strings"

	"github.com/wenlng/go-captcha/captcha"
)

var instance *captcha.Captcha

func getInstance(conf *CaptchaConf) (*captcha.Captcha, error) {
	if instance == nil {
		instance = captcha.GetCaptcha()

		// 设置随机数
		if err := instance.SetRangChars(strings.Split(Chars, "")); err != nil {
			return nil, err
		}
		// 设置字体
		instance.SetFont([]string{conf.FontPath})
		// 设置图片
		instance.SetBackground(conf.Images)
	}

	return instance, nil
}
