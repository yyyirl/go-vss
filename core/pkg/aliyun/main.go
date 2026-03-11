/**
 * @Author:         yi
 * @Description:    main
 * @Version:        1.0.0
 * @Date:           2022/2/21 11:24 AM
 */
package aliyun

import "skeyevss/core/tps"

type AliClient struct {
	SavePath tps.YamlSavePath
	Conf     tps.YamlAli
	Mode     string
}

func New(conf *AliClient) *AliClient {
	return conf
}

// MediaInfoRecord 音视频获取
type AliMediaInfoRecord struct {
	VAIds []string `validate:"required"`
}

// OssBase64ImageRecord oss base64 上传
type AliOssBase64ImageRecord struct {
	Stream   string `json:"stream" validate:"required"`
	FileName string `json:"filename" validate:"required"`
}

type AliUrlFileUploadItem struct {
	Ext      string `json:"ext" validate:"required" msg:"ext" msgpack:"ext"`                // 主键id
	UniqueId string `json:"uniqueId" validate:"uniqueId" msg:"uniqueId" msgpack:"uniqueId"` // 主键id
	Url      string `json:"url" validate:"url"`
	Size     uint64 `json:"size" validate:"size"`
}

// oss url 文件上传
type AliUrlFileRecords struct {
	Records []*AliUrlFileUploadItem `json:"records" validate:"required,dive"`
}

// 媒体裁剪
type AliMediaClippingRecord struct {
	VideoId     string `json:"video_id" validate:"required"`
	Start       int    `json:"start" validate:"required"`
	End         int    `json:"end" validate:"required"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
	Tags        string `json:"tags"`
	CateId      string `json:"cate_id"`
}

// sts 鉴权
type AliOssStsRecord struct {
	Filename string `json:"filename"`
	FType    int    `json:"type"`
	Folder   string `json:"folder"`

	MemberId int64 `json:"memberId,optional"`
}

// 视频
type AliVideoRecord struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	FileName    string `json:"filename" validate:"required"`
	CoverURL    string `json:"cover_url"`
}
