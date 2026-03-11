/**
 * @Author:         yi
 * @Description:    types
 * @Version:        1.0.0
 * @Date:           2022/8/2 16:35
 */
package imgcaptcha

import (
	"context"
	"github.com/wenlng/go-captcha/captcha"
)

type Coordinates struct {
	MinX int
	MinY int

	X int `json:"x"`
	Y int `json:"y"`
}

type (
	Dots map[int]captcha.CharDot

	Transfer struct {
		Expire int64
		Dots   Dots
	}

	Response struct {
		Chars []string    `json:"chars"`
		Image string      `json:"image"`
		Key   string      `json:"key"`
		Dots  interface{} `json:"dots,optional,omitempty"`
	}

	CaptchaConf struct {
		Ctx context.Context
		// 秘钥
		Key string
		// 过期时间
		Expire int64

		ImagePath string
		Images    []string
		FontPath  string
		// 图片宽高
		Width,
		Height int
		// 压缩级别
		QualityCompress int
		// 字数范围
		CharCount,
		// 字体大小范围
		FontSize captcha.RangeVal

		// 字体颜色值
		Colors []string
		// 字体透明度
		FontAlpha float64
		// 字体阴影
		TextShadow bool
		// 字体阴影颜色
		TextShadowColor string
		// 阴影偏移位置
		TextShadowPointX,
		TextShadowPointY int
		// 验证码文本的旋转角度
		TextRangAnglePos []captcha.RangeVal
	}
)
