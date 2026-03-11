/**
 * @Author: 		yi
 * @Description:	oss 上传
 * @Version: 		1.0.0
 * @Date: 			2021-1-17 0:19
 */
package aliyun

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/vod"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-basic/uuid"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/functions/sc"
)

/**
 * @Description: base64上传
 * @param data
 * @return (interface{}, int, error)
 */
func (this *AliClient) Base64ImageUpload(data *AliOssBase64ImageRecord) (interface{}, error) {
	// 保存到本地
	fileName, fullPath, err := functions.MakeBase64Image(this.SavePath.Image, data.Stream)
	if err != nil {
		return nil, err
	}

	// 初始化oss 客户端
	client, err := this.OssClient()
	if err != nil {
		return nil, err
	}

	// bucket
	bucket, err := client.Bucket(this.Conf.Bucket)
	if err != nil {
		return nil, err
	}

	// 上传到oss
	var ossPath string
	if data.FileName != "" {
		ossPath = "images/base64/" + strings.TrimLeft(data.FileName, "/")
	} else {
		ossPath = "images/base64/" + fileName
	}

	fullPath, err = filepath.Abs(fullPath)
	if err != nil {
		return nil, err
	}

	err = bucket.PutObjectFromFile(ossPath, fullPath)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = os.Remove(fullPath)
	}()

	return map[string]interface{}{
		"path": this.Conf.Host + "/" + ossPath,
	}, nil
}

// 本地文件上传
func (this *AliClient) FileUpload(fullPath, fileName string, host bool) (string, error) {
	// 初始化oss 客户端
	client, err := this.OssClient()
	if err != nil {
		return "", err
	}

	// bucket
	bucket, err := client.Bucket(this.Conf.Bucket)
	if err != nil {
		return "", err
	}

	// 上传到oss
	ossPath := "files/web/" + fileName

	err = bucket.PutObjectFromFile(ossPath, fullPath)
	if err != nil {
		return "", err
	}

	if host {
		return this.Conf.Host + "/" + ossPath, err
	}

	return "/" + ossPath, nil
}

// url上传至阿里云
func (this *AliClient) UrlFileUpload(data *AliUrlFileRecords) ([]*AliUrlFileUploadItem, error) {
	if len(data.Records) <= 0 {
		return nil, errors.New("records 不能为空")
	}

	var (
		uniqueId      = functions.UniqueId()
		savePath, err = filepath.Abs(this.SavePath.File + "/" + uniqueId)
	)

	if err != nil {
		return nil, err
	}

	if ok, _ := functions.PathExists(savePath); !ok {
		// 创建目录
		if err := functions.MakeDir(savePath); err != nil {
			return nil, err
		}
	}

	var (
		uploadList []string
		ossPath    = "/images/crop/" + uniqueId + "/"
	)
	for _, item := range data.Records {
		var (
			fileName = uuid.New() + "." + item.Ext // 文件名
			fullPath = savePath + "/" + fileName   // 本地路径
		)

		if err := functions.DownloadFile(item.Url, fullPath); err != nil {
			continue
		}

		uploadList = append(uploadList, fullPath)
		item.Url = ossPath + fileName
		item.Size = functions.FileSize(fullPath)
	}

	if len(uploadList) <= 0 {
		return nil, err
	}

	// 删除源文件
	defer func() {
		_ = os.RemoveAll(savePath)
	}()

	// 上传阿里云 https://www.alibabacloud.com/help/zh/object-storage-service/latest/cp-copy-objects#concept-1937460
	var (
		stdout,
		stderr bytes.Buffer
		cmdString = this.Conf.OssUtilCmd + " cp -r " + savePath +
			" oss://" + this.Conf.Bucket + ossPath +
			"-i " + this.Conf.AccessId +
			" -k " + this.Conf.AccessKey +
			" -e " + this.Conf.Endpoint
	)

	cmd := sc.ExecCommand(cmdString)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		functions.LogError(fmt.Sprintf("\n pdf url 文件上传失败 err: %+v \n cmd: %s \n", err, cmdString))
		return nil, err
	}

	return data.Records, nil
}

// https://help.aliyun.com/document_detail/101411.html
// https://www.alibabacloud.com/help/zh/doc-detail/61388.htm
// 上传音视频点播
func (this *AliClient) UploadMediaToVAServer(fileName, fullPath string) (string, error) {
	client, err := this.InitVodClient()
	if err != nil {
		return "", err
	}

	request := vod.CreateCreateUploadVideoRequest()

	request.Title = fileName
	// request.Description = "Sample Description"
	request.AcceptFormat = "JSON"
	request.FileName = fullPath
	// //request.CateId = "-1"
	// request.CoverURL = "http://192.168.0.0/16/tps/TB1qnJ1PVXXXXXCXXXXXXXXXXXX-700-700.png"
	// request.Tags = "tag1,tag2"

	response, err := client.CreateUploadVideo(request)
	if err != nil {
		return "", err
	}

	// 上传到oss
	var (
		uploadAuthDTO          UploadAuthDTO
		uploadAddressDTO       UploadAddressDTO
		uploadAuthDecode, _    = base64.StdEncoding.DecodeString(response.UploadAuth)
		uploadAddressDecode, _ = base64.StdEncoding.DecodeString(response.UploadAddress)
	)

	err = functions.JSONUnmarshal(uploadAuthDecode, &uploadAuthDTO)
	if err != nil {
		return "", err
	}

	err = functions.JSONUnmarshal(uploadAddressDecode, &uploadAddressDTO)
	if err != nil {
		return "", err
	}

	// 使用UploadAuth和UploadAddress初始化OSS客户端
	ossClient, err := this.VAOssClient(&uploadAuthDTO, &uploadAddressDTO)
	if err != nil {
		return "", err
	}

	if err = this.uploadLocalFile(ossClient, uploadAddressDTO, fullPath); err != nil {
		return "", err
	}

	return response.VideoId, nil
}

// uploadLocalFile 上传本地文件
func (this *AliClient) uploadLocalFile(client *oss.Client, uploadAddressDTO UploadAddressDTO, localFile string) error {
	// 获取存储空间。
	bucket, err := client.Bucket(uploadAddressDTO.Bucket)
	if err != nil {
		return err
	}

	// 上传本地文件。
	err = bucket.PutObjectFromFile(uploadAddressDTO.FileName, localFile)
	if err != nil {
		return err
	}

	return nil
}
