/**
 * @Author:         yi
 * @Description:    pattern
 * @Version:        1.0.0
 * @Date:           2022/10/13 15:19
 */
package functions

import (
	"errors"
	"regexp"
)

type Pattern struct {
	V string
}

// Patterns
var Patterns = map[string]string{
	"backgroundImage": `background\-image\:url\((.+?)\)`,
	"mobile":          "^(((13[0-9])|(14[0-9])|(15[0-9])|(16[0-9])|(17[0-9])|(18[0-9])|(19[0-9]))+\\d{8})$",
	"url":             "((http|ftp|https)://)(([a-zA-Z0-9\\._-]+\\.[a-zA-Z]{2,6})|([0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}))(:[0-9]{1,4})*(/[a-zA-Z0-9\\&%_\\./-~-]*)?",
	"username":        "^[a-zA-Z0-9_-]{4,16}$",
	"email":           "^([a-zA-Z0-9]+[-_|\\_|\\.]?)*[a-zA-Z0-9]+@([a-zA-Z0-9]+[-_|\\_|\\.]?)*[a-zA-Z0-9]+\\.[a-zA-Z]{2,3}$",
	"number":          "^\\d+$",
	"edu":             "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.+-]+((\\.edu)|(\\.edu\\.cn))$",
}

// 初始化
func NewPattern(v string) *Pattern {
	return &Pattern{
		V: v,
	}
}

// 验证
func (p Pattern) match(regular string) bool {
	reg := regexp.MustCompile(regular)
	return reg.MatchString(p.V)
}

// 验证是否是数字
func (p Pattern) IsNaN() bool {
	return !p.match(Patterns["number"])
}

// 验证是否是手机号码
func (p Pattern) IsMobile() bool {
	return p.match(Patterns["mobile"])
}

// 验证是否是邮箱
func (p Pattern) IsEmail() bool {
	return p.match(Patterns["email"])
}

// 验证用户名
func (p Pattern) IsUsername() bool {
	return p.match(Patterns["username"])
}

// 验证密码
func (p Pattern) IsPassword() bool {
	length := len(p.V)
	return length <= 20 && length >= 6
}

// 是否是url
func (p Pattern) IsUrl() bool {
	return p.match(Patterns["url"])
}

// 是否是edu邮箱
func (p Pattern) IsEduMail() bool {
	return p.match(Patterns["edu"])
}

func ReplaceStringByRegex(str, rule, replace string) (string, error) {
	reg, err := regexp.Compile(rule)
	if reg == nil || err != nil {
		return "", errors.New("正则MustCompile错误:" + err.Error())
	}
	return reg.ReplaceAllString(str, replace), nil
}
