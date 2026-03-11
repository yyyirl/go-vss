/**
 * @Author:         yi
 * @Description:    email
 * @Version:        1.0.0
 * @Date:           2025/1/13 18:15
 */
package common

import (
	"errors"
	"time"

	"skeyevss/core/constants"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/response"
	"skeyevss/core/tps"
)

type Email struct {
	CacheExpire int64
	AesKey      string
	Config      tps.YamlEmail
}

func NewEmail(CacheExpire int64, AesKey string, Config tps.YamlEmail) *Email {
	return &Email{
		CacheExpire: CacheExpire,
		AesKey:      AesKey,
		Config:      Config,
	}
}

func (e *Email) EmailCaptcha(name, email, Type string) (string, *response.HttpErr) {
	if !functions.NewPattern(email).IsEmail() {
		return "", response.MakeError(response.NewHttpRespMessage().Str("邮箱格式错误"), localization.M0001)
	}

	var (
		title,
		subject,
		body string
		code = functions.GetCode()
	)
	switch Type {
	case constants.EMAIL_KEY_REGISTER:
		title = "登录 - " + name
		subject = "登录验证码"
		body = "您的验证码为 " + code

	case constants.EMAIL_KEY_LOGIN:
		title = "注册 - " + name
		subject = "注册验证码"
		body = "您的验证码为 " + code

	case constants.EMAIL_KEY_FIND_PWD:
		title = "找回密码 - " + name
		subject = "找回密码验证码"
		body = "您的验证码为 " + code

	case constants.EMAIL_KEY_BIND:
		title = "验证码 - " + name
		subject = "绑定验证码"
		body = "您的验证码为 " + code

	case constants.EMAIL_KEY_UNBIND:
		title = "验证码 - " + name
		subject = "解绑验证码"
		body = "您的验证码为 " + code

	case constants.EMAIL_KEY_BIND_BANK_ACCOUNT:
		title = "验证码 - " + name
		subject = "绑定银行卡"
		body = "您的验证码为 " + code

	default:
		return "", response.MakeError(response.NewHttpRespMessage().Str("发送类型错误"), localization.M0001)
	}

	// 验证码失效时间
	d, err := functions.JSONMarshal(tps.EmailDecrypt{
		Email:  email,
		Code:   code,
		Expire: functions.NewTimer().Now() + e.CacheExpire,
		Type:   Type,
	})

	if err != nil {
		return "", response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00271)
	}

	// 生成秘钥
	encrypt, err := functions.NewCrypto([]byte(e.AesKey)).Encrypt(d)
	if err != nil {
		return "", response.MakeError(response.NewHttpRespMessage().Err(err), localization.M0003)
	}

	// 发送邮件
	if err := functions.NewMail(
		e.Config.Username,
		e.Config.Password,
		e.Config.Host,
		e.Config.Port,
	).Send(email, title, subject, body); err != nil {
		return "", response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00273)
	}

	return encrypt, nil
}

func (e *Email) VerifyEmailCode(Type, code, emailEncrypt, email string) error {
	// 检测邮箱验证码
	encrypt, err := functions.NewCrypto([]byte(e.AesKey)).Decrypt(emailEncrypt)
	if err != nil {
		return err
	}

	var j tps.EmailDecrypt
	if err := functions.JSONUnmarshal([]byte(encrypt), &j); err != nil {
		return err
	}

	if j.Expire < time.Now().Unix() {
		return errors.New("验证码已过期")
	}

	if j.Email != email {
		return errors.New("邮箱不匹配")
	}

	if j.Code != code {
		return errors.New("验证码不匹配")
	}

	if j.Type != Type {
		return errors.New("类型不匹配")
	}

	return nil
}
