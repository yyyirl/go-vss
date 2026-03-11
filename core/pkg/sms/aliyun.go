/**
 * @Author:         yi
 * @Description:    阿里大于短信
 * @Version:        1.0.0
 * @Date:           2022/10/20 21:40
 */
package sms

import (
	"errors"
	"time"

	"github.com/crazytaxii/aliyuncs"

	"skeyevss/core/tps"
)

func SendMessageWithAliDy(conf *tps.YamlAliSms, mobile, template string, params map[string]string) error {
	if conf == nil {
		return errors.New("参数错误")
	}

	_, err := aliyuncs.NewClient(
		conf.AccessKeyId,
		conf.AccessKeySecret,
		10*time.Second,
	).SendSMS(
		mobile,
		// 短信签名
		conf.SignName,
		// 短信模板变量
		template,
		params,
	)

	return err
}
