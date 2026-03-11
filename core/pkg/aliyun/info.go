/**
 * @Author: 		yi
 * @Description:	信息获取
 * @Version: 		1.0.0
 * @Date: 			2021-2-1 20:44
 * @Example:		https://help.aliyun.com/document_detail/101427.html?spm=a2c4g.11186623.6.1130.6c65d418xqt96m
 */
package aliyun

import (
	"errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vod"
	"strings"
)

type VideoItem struct {
	Title        string
	CateName     string
	Description  string
	CreationTime string
	CoverURL     string
	VideoId      string
	Snapshots    []string
}

// 批量获取视频信息
func (this *AliClient) MediaInfo(videoIds []string) (map[string]*VideoItem, error) {
	if len(videoIds) <= 0 {
		return nil, errors.New("参数错误, vaId不能为空")
	}

	// 创建客户端
	client, err := this.InitVodClient()
	if err != nil {
		return nil, err
	}

	request := vod.CreateGetVideoInfosRequest()
	request.VideoIds = strings.Join(videoIds, ",")
	request.AcceptFormat = "JSON"
	results, err := client.GetVideoInfos(request)
	if err != nil {
		return nil, err
	}

	if len(results.VideoList) <= 0 {
		return nil, errors.New("未查到相关列表")
	}

	var list map[string]*VideoItem
	list = make(map[string]*VideoItem, 0)
	for _, val := range results.VideoList {
		list[val.VideoId] = &VideoItem{
			Title:        val.Title,
			CateName:     val.CateName,
			Description:  val.Description,
			CreationTime: val.CreationTime,
			CoverURL:     val.CoverURL,
			VideoId:      val.VideoId,
			Snapshots:    val.Snapshots,
		}
	}

	return list, nil
}

// 获取播放信息 https://help.aliyun.com/document_detail/101407.html?spm=a2c4g.11186623.6.1129.2fa7d41886oJKW
func (this *AliClient) MediaPlayInfo(id string) (map[string]interface{}, error) {
	if id == "" {
		return nil, errors.New("id不能我空")
	}

	// 创建客户端
	client, err := this.InitVodClient()
	if err != nil {
		return nil, err
	}

	// requests
	request := vod.CreateGetPlayInfoRequest()
	request.VideoId = id
	request.AcceptFormat = "JSON"
	_, _ = client.GetPlayInfo(request)

	response, err := client.GetPlayInfo(request)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"url":      response.PlayInfoList.PlayInfo[0].PlayURL,
		"width":    response.PlayInfoList.PlayInfo[0].Width,
		"height":   response.PlayInfoList.PlayInfo[0].Height,
		"duration": response.PlayInfoList.PlayInfo[0].Duration,
		"bitrate":  response.PlayInfoList.PlayInfo[0].Bitrate,
		"format":   response.PlayInfoList.PlayInfo[0].Format,
		"size":     response.PlayInfoList.PlayInfo[0].Size,
		"id":       id,
		"cover":    response.VideoBase.CoverURL,
		"title":    response.VideoBase.Title,
		"snapshot": "",
	}, nil
}

// 获取播放信息
func (this *AliClient) MediaPlayAddress(id string) (interface{}, error) {
	res, err := this.MediaPlayInfo(id)

	return res, err
}
