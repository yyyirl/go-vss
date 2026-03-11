/**
 * @Author:         yi
 * @Description:    green
 * @Version:        1.0.0
 * @Date:           2024/9/25 23:00
 */
package aliyun

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	green20220302 "github.com/alibabacloud-go/green-20220302/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"

	"skeyevss/core/constants"
	"skeyevss/core/pkg/functions"
)

var (
	greenClient     *green20220302.Client
	greenClientConf *openapi.Config
)

const (
	AliGreenTextRiskLevelNone   uint = iota // none 低风险
	AliGreenTextRiskLevelLow                // low 低风险
	AliGreenTextRiskLevelMedium             // medium 中风险
	AliGreenTextRiskLevelHigh               // high 高风险
)

const (
	AliGreenTextRiskLevelNoneResp   = "none"   // none 低风险
	AliGreenTextRiskLevelLowResp    = "low"    // low 低风险
	AliGreenTextRiskLevelMediumResp = "medium" // medium 中风险
	AliGreenTextRiskLevelHighResp   = "high"   // high 高风险
)

var AliGreenTextRiskLevelTypes = []uint{
	AliGreenTextRiskLevelNone,
	AliGreenTextRiskLevelLow,
	AliGreenTextRiskLevelMedium,
	AliGreenTextRiskLevelHigh,
}

const AliGreenTextRiskLevelExplain = "审核拒绝原因"

var AliGreenTextRiskLevels = []*constants.ConfItemInner{
	{
		Id: AliGreenTextRiskLevelNone,
		Title: &constants.Lang{
			ZH: AliGreenTextRiskLevelNoneResp,
			EN: AliGreenTextRiskLevelNoneResp,
		},
	},
	{
		Id: AliGreenTextRiskLevelLow,
		Title: &constants.Lang{
			ZH: AliGreenTextRiskLevelLowResp,
			EN: AliGreenTextRiskLevelLowResp,
		},
	},
	{
		Id: AliGreenTextRiskLevelMedium,
		Title: &constants.Lang{
			ZH: AliGreenTextRiskLevelMediumResp,
			EN: AliGreenTextRiskLevelMediumResp,
		},
	},
	{
		Id: AliGreenTextRiskLevelHigh,
		Title: &constants.Lang{
			ZH: AliGreenTextRiskLevelHighResp,
			EN: AliGreenTextRiskLevelHighResp,
		},
	},
}

// https://help.aliyun.com/document_detail/434034.html?spm=a2c4g.433945.0.0.1f063104yZOsFJ
// 响应检测 https://next.api.aliyun.com/troubleshoot?spm=a2c4g.11186623.0.0.63176694J8CHuC&q=%7B%0A%20%20%20%22Code%22%3A%20400%2C%0A%20%20%20%22Message%22%3A%20%22service%20is%20invalid%22%2C%0A%20%20%20%22RequestId%22%3A%20%2299D9002E-021D-58E3-A610-49BA83A3FCFD%22%0A%7D&requestId=&product=
// 文本检测 https://help.aliyun.com/document_detail/467828.html?spm=a2c4g.467825.0.0.7cbc3104fRK2Tp#065ca2b17cclu
func (this *AliClient) GreenText(level uint, content string) (string, error) {
	if level == AliGreenTextRiskLevelNone {
		return "", nil
	}

	if len(strings.TrimSpace(content)) == 0 {
		return "", nil
	}

	bytesData, err := functions.JSONMarshal(map[string]interface{}{"content": content})
	if err != nil {
		return "", err
	}

	var textModerationRequest = &green20220302.TextModerationPlusRequest{
		// Service:           tea.String("service code"),
		Service:           tea.String("llm_query_moderation"),
		ServiceParameters: tea.String(string(bytesData)),
	}

	// 创建客户端
	if greenClient == nil || greenClientConf == nil {
		greenClient, greenClientConf, err = this.GreenClient()
		if err != nil {
			return "", err
		}
	}

	var (
		runtime      = new(util.RuntimeOptions)
		flag         = false
		parseMessage = func(data *green20220302.TextModerationPlusResponseBodyData) string {
			var messages []string
			for _, item := range data.Result {
				messages = append(messages, fmt.Sprintf("[%s] %s", *item.RiskWords, *item.Description))
			}

			return strings.Join(messages, ", ") + "-" + *data.RiskLevel
		}
	)
	runtime.ReadTimeout = tea.Int(10000)
	runtime.ConnectTimeout = tea.Int(10000)

	// 自动路由，服务端错误，区域切换至cn-beijing。
	response, err := greenClient.TextModerationPlusWithOptions(textModerationRequest, runtime)
	if err != nil {
		var err1 = &tea.SDKError{}
		if v, ok := err.(*tea.SDKError); ok {
			err1 = v
			if *err1.StatusCode == 500 {
				flag = true
			}
		}
	}
	if response == nil || *response.StatusCode == 500 || *response.Body.Code == 500 {
		flag = true
	}

	if flag {
		greenClientConf.SetRegionId("cn-beijing")
		greenClientConf.SetEndpoint("green-cip.cn-beijing.aliyuncs.com")
		greenClient, err := green20220302.NewClient(greenClientConf)
		if err != nil {
			return "", err
		}

		response, err = greenClient.TextModerationPlusWithOptions(textModerationRequest, runtime)
		if err != nil {
			return "", err
		}
	}

	if response != nil {
		var (
			statusCode = tea.IntValue(tea.ToInt(response.StatusCode))
			body       = response.Body
		)

		if statusCode == http.StatusOK {
			if tea.IntValue(tea.ToInt(body.Code)) == 200 {
				if body.Data.RiskLevel != nil {
					var riskLevel = *body.Data.RiskLevel
					if riskLevel == "none" {
						return "", nil
					}

					if riskLevel == AliGreenTextRiskLevelLowResp {
						if functions.Contains(level, []uint{AliGreenTextRiskLevelMedium, AliGreenTextRiskLevelHigh}) {
							return "", nil
						}

						return parseMessage(body.Data), nil
					}

					if riskLevel == AliGreenTextRiskLevelMediumResp {
						if functions.Contains(level, []uint{AliGreenTextRiskLevelHigh}) {
							return "", nil
						}

						return parseMessage(body.Data), nil
					}

					if riskLevel == AliGreenTextRiskLevelHighResp {
						return parseMessage(body.Data), nil
					}
				}

				return "", nil
			}

			return "", fmt.Errorf(body.String())
		} else {
			return "", fmt.Errorf("response not success. status:" + tea.ToString(statusCode))
		}
	}
	return "", errors.New("响应为空")
}
