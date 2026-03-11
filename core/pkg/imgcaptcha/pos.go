/**
 * @Author:         yi
 * @Description:    pos
 * @Version:        1.0.0
 * @Date:           2023/9/15 16:34
 */
package imgcaptcha

import (
	"github.com/wenlng/go-captcha/captcha"
	"golang.org/x/image/font"

	"skeyevss/core/constants"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/functions"
)

type CaptchaPos struct {
	conf *CaptchaConf
}

// https://github.com/wenlng/go-captcha/blob/master/README_zh.md
func NewCaptchaPos(conf CaptchaConf) *CaptchaPos {
	return &CaptchaPos{
		conf: &conf,
	}
}

func (c *CaptchaPos) Make() (*Response, error) {
	if len(c.conf.Images) <= 0 {
		images, err := functions.GetAllFilesInFolder(c.conf.ImagePath)
		if err != nil {
			return nil, err
		}
		c.conf.Images = images
	}

	instance, err := getInstance(c.conf)
	if err != nil {
		return nil, err
	}
	// 设置验证码主图的尺寸
	instance.SetImageSize(captcha.Size{c.conf.Width, c.conf.Height})
	// 设置验证码主图清晰度，压缩级别范围 QualityCompressLevel1 - 5，QualityCompressNone无压缩，默认为最低压缩级别
	instance.SetImageQuality(c.conf.QualityCompress)
	// 设置字体Hinting值 (HintingNone,HintingVertical,HintingFull)
	instance.SetFontHinting(font.HintingFull)
	// 设置验证码文本显示的总数随机范围
	instance.SetTextRangLen(c.conf.CharCount)
	// 设置验证码文本的随机大小
	instance.SetRangFontSize(c.conf.FontSize)
	// 设置验证码文本的随机十六进制颜色
	instance.SetTextRangFontColors(c.conf.Colors)
	// 字体的透明度
	instance.SetImageFontAlpha(c.conf.FontAlpha)
	// 设置字体阴影
	instance.SetTextShadow(c.conf.TextShadow)
	if c.conf.TextShadow {
		// 设置字体阴影颜色
		if c.conf.TextShadowColor != "" {
			instance.SetTextShadowColor(c.conf.TextShadowColor)
		}
		// 设置字体阴影偏移位置
		if c.conf.TextShadowPointX != 0 && c.conf.TextShadowPointY != 0 {
			instance.SetTextShadowPoint(captcha.Point{c.conf.TextShadowPointX, c.conf.TextShadowPointY})
		}
	}
	// 验证码文本的旋转角度
	instance.SetTextRangAnglePos(c.conf.TextRangAnglePos)
	// 验证码字体的扭曲程度
	instance.SetImageFontDistort(captcha.DistortLevel2)

	// 生成验证码
	dots, b64, _, _, err := instance.Generate()
	if err != nil {
		return nil, err
	}

	key, err := DotsEncryption(c.conf.Key, functions.NewTimer().Now()+c.conf.Expire, dots)
	if err != nil {
		return nil, err
	}

	var chars []string
	for _, item := range dots {
		chars = append(chars, item.Text)
	}

	var data = &Response{
		Image: b64,
		Key:   key,
		Chars: chars,
	}
	if contextx.GetReqDev(c.conf.Ctx) == constants.ENV_DEVELOPMENT {
		data.Dots = dots
	}

	return data, nil
}
