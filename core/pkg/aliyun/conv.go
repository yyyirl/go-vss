/**
 * @Author:         yi
 * @Description:    conv
 * @Version:        1.0.0
 * @Date:           2022/1/17 18:52
 */
package aliyun

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vod"
)

func (this *AliClient) AudioConv(vaId string) (interface{}, error) {
	// 创建客户端
	client, err := this.InitVodClient()
	if err != nil {
		return nil, err
	}

	request := vod.CreateSubmitTranscodeJobsRequest()
	request.VideoId = vaId
	request.TemplateGroupId = "09c47b9d94526286f309b2d71a4bd330"
	request.AcceptFormat = "JSON"

	response, err := client.SubmitTranscodeJobs(request)

	if err != nil {
		panic(err)
	}

	for _, job := range response.TranscodeJobs.TranscodeJob {
		fmt.Printf("%s\n", job.JobId)
	}
	return nil, nil
}
