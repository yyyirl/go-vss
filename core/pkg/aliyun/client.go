/**
 * @Author: 		yi
 * @Description:	阿里云 Vod Client
 * @Version: 		1.0.0
 * @Date: 			2021-1-16 20:20
 */
package aliyun

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	green20220302 "github.com/alibabacloud-go/green-20220302/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vod"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type UploadAuthDTO struct {
	AccessKeyId,
	AccessKeySecret,
	SecurityToken string
}

type UploadAddressDTO struct {
	Endpoint,
	Bucket,
	FileName string
}

// 初始化
func (this *AliClient) InitVodClient() (client *vod.Client, err error) {
	// 自定义config
	config := sdk.NewConfig()
	// 失败是否自动重试
	config.AutoRetry = true
	// 最大重试次数
	config.MaxRetryTime = 3
	// 连接超时，单位：纳秒；默认为3秒
	config.Timeout = 3000000000

	// 创建vodClient实例
	return vod.NewClientWithOptions(
		// 点播服务接入区域
		this.Conf.RegionIdVideo,
		config,
		// 授权对象
		&credentials.AccessKeyCredential{
			AccessKeyId:     this.Conf.AccessId,
			AccessKeySecret: this.Conf.AccessKey,
		},
	)
}

// oss client
func (this *AliClient) OssClient() (*oss.Client, error) {
	var endpoint = this.Conf.Endpoint
	if this.Mode == "pro" {
		// 内网
		endpoint = this.Conf.EndpointIntranet
	}

	return oss.New(
		endpoint,
		this.Conf.AccessId,
		this.Conf.AccessKey,
	)
}

// 音视频点播oss上传
func (this *AliClient) VAOssClient(uploadAuthDTO *UploadAuthDTO, uploadAddressDTO *UploadAddressDTO) (*oss.Client, error) {
	return oss.New(
		uploadAddressDTO.Endpoint,
		uploadAuthDTO.AccessKeyId,
		uploadAuthDTO.AccessKeySecret,
		oss.SecurityToken(uploadAuthDTO.SecurityToken),
		oss.Timeout(86400*7, 86400*7),
	)
}

// 内容审核
func (this *AliClient) GreenClient() (*green20220302.Client, *openapi.Config, error) {
	var config = &openapi.Config{
		AccessKeyId:     tea.String(this.Conf.AccessId),
		AccessKeySecret: tea.String(this.Conf.AccessKey),
		// RegionId: tea.String("cn-shanghai"),
		RegionId:       tea.String(this.Conf.RegionId),
		Endpoint:       tea.String("green-cip." + this.Conf.RegionId + ".aliyuncs.com"),
		ConnectTimeout: tea.Int(3000),
		ReadTimeout:    tea.Int(6000),
	}

	client, err := green20220302.NewClient(config)
	return client, config, err
}
