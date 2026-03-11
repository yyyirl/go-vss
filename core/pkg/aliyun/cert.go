/**
 * @Author: 		yi
 * @Description:	阿里云上传鉴权
 * @Version: 		1.0.0
 * @Date: 			2021-1-16 20:40
 */
package aliyun

import (
	"strconv"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/vod"
	"github.com/aliyun/aliyun-sts-go-sdk/sts"

	"skeyevss/core/pkg/functions"
)

// sts token (oss) https://github.com/aliyun/aliyun-sts-go-sdk
func (this *AliClient) StsFile(data *AliOssStsRecord) (interface{}, error) {
	resp, err := sts.NewClient(
		this.Conf.StsAccessId,
		this.Conf.StsAccessKey,
		this.Conf.StsSetRoleArn,
		this.Conf.StsRoleName,
	).AssumeRole(uint(this.Conf.StsTokenExpire))
	if err != nil {
		return nil, err
	}

	var folder = "/known"
	data.Folder = "/" + strings.TrimSpace(strings.Trim(data.Folder, "/"))
	if data.Folder != "" {
		folder = data.Folder
	} else {
		if data.MemberId != 0 {
			folder = "/mb-" + strconv.FormatInt(data.MemberId, 10)
		}
	}

	return map[string]interface{}{
		"endpoint":   "https://" + this.Conf.Endpoint,
		"uploadHost": "https://" + this.Conf.Bucket + "." + this.Conf.RegionId + ".aliyuncs.com",
		// "previewHost": "https://" + this.Conf.Bucket + ".oss-" + this.Conf.RegionId + ".aliyuncs.com",
		"previewHost": this.Conf.Host,
		// "type":        "scs",
		"ossParams": map[string]interface{}{
			"accessKeyId":     resp.Credentials.AccessKeyId,
			"accessKeySecret": resp.Credentials.AccessKeySecret,
			"expiration":      resp.Credentials.Expiration.Unix(),
			// "securityToken":   resp.Credentials.SecurityToken,
			"stsToken": resp.Credentials.SecurityToken,
			"region":   "oss-" + this.Conf.RegionId,
			"secure":   true,
			"bucket":   this.Conf.Bucket,
			"fileName": folder + Folder(data.FType) + functions.FileRename(data.Filename),
		},
	}, nil
}

// sts token (oss) https://github.com/aliyun/aliyun-sts-go-sdk
func (this *AliClient) StsToken() (interface{}, error) {
	resp, err := sts.NewClient(
		this.Conf.StsAccessId,
		this.Conf.StsAccessKey,
		this.Conf.StsSetRoleArn,
		this.Conf.StsRoleName,
	).AssumeRole(uint(this.Conf.StsTokenExpire))
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"endpoint":    "http://" + this.Conf.Endpoint,
		"uploadHost":  "https://" + this.Conf.Bucket + "." + this.Conf.RegionId + ".aliyuncs.com",
		"previewHost": this.Conf.Host,
		"ossParams": map[string]interface{}{
			"accessKeyId":     resp.Credentials.AccessKeyId,
			"accessKeySecret": resp.Credentials.AccessKeySecret,
			"expiration":      resp.Credentials.Expiration.Unix(),
			"stsToken":        resp.Credentials.SecurityToken,
			"region":          "oss-" + this.Conf.RegionId,
			"secure":          true,
			"bucket":          this.Conf.Bucket,
		},
	}, nil
}

// Folder oss 文件夹
func Folder(fType int) string {
	switch fType {
	case 1:
		return "/images/"
	case 2:
		return "/videos/"
	case 3:
		return "/document/pdf/"
	case 4:
		return "/document/word/"
	case 5:
		return "/document/txt/"
	default:
		return "/unknown/"
	}
}

// VideoToken 视频点播上传凭证 https://help.aliyun.com/document_detail/101411.html
func (this *AliClient) VideoToken(data *AliVideoRecord) (interface{}, error) {
	// 创建客户端
	client, err := this.InitVodClient()
	if err != nil {
		return nil, err
	}

	// 音视频上传信息
	request := vod.CreateCreateUploadVideoRequest()
	request.Title = data.Title
	request.Description = data.Description
	request.FileName = data.FileName
	// request.CateId = "-1"
	// request.Tags = "tag1,tag2"
	request.CoverURL = data.CoverURL
	request.AcceptFormat = "JSON"

	response, err := client.CreateUploadVideo(request)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"VideoId":       response.VideoId,
		"UploadAddress": response.UploadAddress,
		"RequestId":     response.RequestId,
		"UploadAuth":    response.UploadAuth,
		"originTitle":   data.Title,
	}, nil
}
