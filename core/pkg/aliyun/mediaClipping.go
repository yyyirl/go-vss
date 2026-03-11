/**
 * @Author: 		yi
 * @Description:	媒体处理
 * @Version: 		1.0.0
 * @Date: 			2021-1-18 20:08
 */
package aliyun

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vod"

	"skeyevss/core/pkg/functions"
)

// 视频点播 音视频裁剪 https://help.aliyun.com/document_detail/101580.html?spm=a2c4g.11186623.6.1136.32f743deXPCPhq
func (this *AliClient) MediaClipping(data *AliMediaClippingRecord) (interface{}, error) {
	// 创建客户端
	client, err := this.InitVodClient()
	if err != nil {
		return nil, err
	}

	// timeline
	request := vod.CreateProduceEditingProjectVideoRequest()
	jsonTimeline, err1 := functions.JSONMarshal(this.makeVideoTracksTimeLine(data))
	if err1 != nil {
		return nil, err
	}
	request.Timeline = string(jsonTimeline)

	// media metadata
	jsonMeta, err := functions.JSONMarshal(map[string]interface{}{
		"Title":       data.Title,
		"Description": data.Description,
		"CoverURL":    data.CoverURL,
		"Tags":        data.Tags,
		"CateId":      data.CateId,
	})

	if err != nil {
		return nil, err
	}
	request.MediaMetadata = string(jsonMeta)
	request.AcceptFormat = "JSON"

	response, err := client.ProduceEditingProjectVideo(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (this *AliClient) makeVideoTracksTimeLine(data *AliMediaClippingRecord) map[string]interface{} {
	// set timeline, this sample shows how to merge two videos
	var videoTracks []map[string]interface{}
	var videoTrackClips []map[string]interface{}
	videoTrackClips = append(videoTrackClips, map[string]interface{}{
		"MediaId": data.VideoId,
		"In":      data.Start,
		"Out":     data.End,
	})
	videoTrack := map[string]interface{}{"VideoTrackClips": videoTrackClips}
	videoTracks = append(videoTracks, videoTrack)

	return map[string]interface{}{"VideoTracks": videoTracks}
}
