/**
 * @Author:         yi
 * @Description:    email
 * @Version:        1.0.0
 * @Date:           2022/10/11 16:06
 */
package functions

import (
	"errors"
	"strconv"
	"strings"
	"time"

	gomail "gopkg.in/gomail.v2"

	"skeyevss/core/tps"
)

type Email struct {
	user,
	pass,
	host,
	port string
}

func NewMail(user, pass, host, port string) *Email {
	return &Email{
		user: user,
		pass: pass,
		host: host,
		port: port,
	}
}

/**
 * @Description: 邮件发送
 * @param mails 邮件
 * @param title 标题
 * @param subject 邮件主题
 * @param body 邮件内容
 * @return error
 * @example https://www.cnblogs.com/fanbi/p/11490241.html
 *   	//定义收件人
 *		err := pkg.SendMail(
 *			[]string{"1003275805@qq.com"},
 *			"我来测试",
 *			"subjectsubjectsubject 测试邮件发送",
 *			"bodybodybody 测试邮件发送测试邮件发送测试邮件发送测试邮件发送",
 *		)
 */
func (m *Email) Send(mails string, title string, subject string, body string) error {
	if mails == "" {
		return nil
	}

	var mailList []string
	for _, item := range strings.Split(mails, ",") {
		mailList = append(mailList, strings.TrimSpace(item))
	}

	if len(mailList) <= 0 {
		return nil
	}

	mailList = ArrUnique(mailList)
	if len(mailList) <= 0 {
		return errors.New("emails 不能为空")
	}

	var mail = gomail.NewMessage()
	mail.SetHeader(
		"From",
		mail.FormatAddress(
			m.user,
			title,
		),
	)
	// 发送给多个用户
	mail.SetHeader("To", mailList...)
	// 设置邮件主题
	mail.SetHeader("Subject", subject)
	mail.SetHeader("Content-Type", "text/html; charset=UTF-8")
	// 设置邮件正文
	mail.SetBody("text/html", body)

	port, _ := strconv.Atoi(m.port)
	return (gomail.NewDialer(
		m.host,
		port,
		m.user,
		m.pass,
	)).DialAndSend(mail)
}

func VerifyEmailCode(Type, code, emailEncrypt, aesKey, email string) error {
	// 检测邮箱验证码
	encrypt, err := NewCrypto([]byte(aesKey)).Decrypt(emailEncrypt)
	if err != nil {
		return err
	}

	var j tps.EmailDecrypt
	if err := JSONUnmarshal([]byte(encrypt), &j); err != nil {
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
