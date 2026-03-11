package middleware

import (
	"strings"

	"skeyevss/core/app/sev/backend/internal/config"
	"skeyevss/core/pkg/functions"
)

func recoverCallback(c config.Config) func(info, broken string) {
	// TODO 改成写日志
	return func(info, broken string) {
		// 发送邮件
		if err := functions.NewMail(
			c.Email.Username,
			c.Email.Password,
			c.Email.Host,
			c.Email.Port,
		).Send(
			c.Email.Emails,
			c.Name+" api application error",
			c.Name+" api application panic error",
			"<p style=\"color:red\">"+info+"</p>"+strings.ReplaceAll(broken, "\n", "<br/>"),
		); err != nil {
			functions.LogError("邮件发送失败 err: ", err)
		}
	}
}
