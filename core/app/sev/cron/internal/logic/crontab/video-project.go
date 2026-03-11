// @Title        录像计划
// @Description  main
// @Create       yiyiyi 2025/7/8 11:28

package crontab

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"skeyevss/core/app/sev/cron/internal/types"
	"skeyevss/core/app/sev/db/client/backendservice"
	"skeyevss/core/app/sev/db/client/configservice"
	"skeyevss/core/app/sev/db/client/deviceservice"
	"skeyevss/core/app/sev/db/pkg/conv"
	"skeyevss/core/common/stream"
	cTypes "skeyevss/core/common/types"
	"skeyevss/core/common/videoProject"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/pkg/response"
	"skeyevss/core/repositories/models/channels"
	"skeyevss/core/repositories/models/crontab"
	videoProjects "skeyevss/core/repositories/models/video-projects"
)

var (
	_ types.CrontabLogic = (*VideoProjectLogic)(nil)

	makeRecord sync.Once

	videoProjectRedisLockKey = "video-project-record-cron-lock"
)

type VideoProjectLogic struct {
	executing bool

	StartRecordingIds,
	StopRecordingIds chan map[uint64]*cTypes.ChannelMSRelItem

	execChannelUniqueIds []uint64
}

func (l *VideoProjectLogic) Executing() bool {
	return l.executing
}

func (l *VideoProjectLogic) Key() string {
	return crontab.UniqueIdVideoProject
}

func (l *VideoProjectLogic) DO(params *types.CrontabLogicDOParams) {
	defer params.Recover("录像计划[1]")

	makeRecord.Do(func() {
		go l.makeRecord(params)
	})

	if params.CrontabRecord.BlockStatus == 0 {
		go l.do(params)
		return
	}

	l.do(params)
}

func (l *VideoProjectLogic) logs(data []string, item string) []string {
	var records = append(data, fmt.Sprintf("[%s] err: %s", functions.NewTimer().Format(""), item))
	if len(records) < 100 {
		return records
	}

	// 保留最后100条
	return append([]string{}, records[len(records)-100:]...)
}

func (l *VideoProjectLogic) do(params *types.CrontabLogicDOParams) {
	l.executing = true

	defer func() {
		l.executing = false
		params.Recover("流媒体保活[2]")
	}()

	if ok, err := params.SvcCtx.RedisClient.AcquireLock(videoProjectRedisLockKey, params.SvcCtx.Config.InternalIP, 60*time.Second); err != nil {
		var message = fmt.Sprintf("!!!录像计划分布式锁获取失败, line: %s; err: %s", functions.Caller(1), err.Error)
		params.CrontabRecord.Logs = l.logs(params.CrontabRecord.Logs, message)
		functions.LogError("录像计划[1]", message)
		return
	} else {
		if !ok {
			return
		}
	}

	defer func() {
		if _, err := params.SvcCtx.RedisClient.ReleaseLock(videoProjectRedisLockKey, params.SvcCtx.Config.InternalIP); err != nil {
			var message = fmt.Sprintf("!!!录像计划分布式锁释放失败, line: %s; err: %s", functions.Caller(1), err.Error)
			params.CrontabRecord.Logs = l.logs(params.CrontabRecord.Logs, message)
			functions.LogError("录像计划[2]", message)
			return
		}
	}()

	// 获取计划任务
	res, err := response.NewRpcToHttpResp[*configservice.Response, *response.ListResp[[]*videoProjects.Item]]().Parse(
		func() (*configservice.Response, error) {
			data, err := conv.New(params.SvcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
				All: true,
				Conditions: []*orm.ConditionItem{
					{Column: videoProjects.ColumnState, Value: 1},
					{Column: videoProjects.ColumnChannelUniqueIds, Value: "[]", Operator: "!="},
				},
			})
			if err != nil {
				return nil, err
			}

			return params.SvcCtx.RpcClients.Config.VideoProjectList(params.Ctx, data)
		},
	)
	if err != nil {
		var message = fmt.Sprintf("录像计划获取失败, line: %s; err: %s", functions.Caller(1), err.Error)
		params.CrontabRecord.Logs = l.logs(params.CrontabRecord.Logs, message)
		functions.LogError("录像计划[3]", message)
		return
	}

	var (
		hourStep   = 24
		channelIds []uint64
		weekNum    = functions.NewTimer().TimestampWeekDay(params.Now)
	)
	for _, item := range res.Data.List {
		// 检测计划时间
		if item.Plans == "" {
			continue
		}

		var plains = strings.Split(item.Plans, "")
		if len(plains) != 168 {
			continue
		}

		var weekMaps = make(map[int][24]string)
		for i := 1; i <= 7; i++ {
			weekMaps[i] = [24]string(plains[(i-1)*hourStep : hourStep*i])
		}

		hourState, ok := weekMaps[weekNum]
		if !ok {
			continue
		}

		var ranges = stream.NewVideoPlain().GetTimeRanges(hourState, params.Now)
		if len(ranges) <= 0 {
			continue
		}

		var exists = false
	LoopInner:
		for _, val := range ranges {
			if params.Now >= val[0] && params.Now <= val[1] {
				exists = true
				break LoopInner
			}
		}

		if !exists {
			continue
		}
		channelIds = append(channelIds, item.ChannelUniqueIds...)
	}

	// 开始/继续录像
	channelIds = functions.ArrUnique(channelIds)
	// 停止录像
	var stopChannelIds []uint64
	for _, item := range l.execChannelUniqueIds {
		if !functions.Contains(item, channelIds) {
			stopChannelIds = append(stopChannelIds, item)
		}
	}

	var queryChannelIds = functions.SliceToSliceAny(append(channelIds, stopChannelIds...))
	if len(queryChannelIds) <= 0 {
		return
	}

	// 获取通道
	channelMaps, err := response.NewRpcToHttpResp[*deviceservice.Response, map[uint64]*cTypes.ChannelMSRelItem]().Parse(
		func() (*backendservice.Response, error) {
			data, err := conv.New(params.SvcCtx.Config.Mode).ToPBParams(&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: channels.ColumnID, Values: queryChannelIds},
				},
				All: true,
			})
			if err != nil {
				return nil, response.NewMakeRpcRetErr(err, 2)
			}

			return params.SvcCtx.RpcClients.Device.MediaServersWithChannelIds(params.Ctx, data)
		},
	)
	if err != nil {
		var message = fmt.Sprintf("录像计划通道获取失败, line: %s; err: %s", functions.Caller(1), err.Error)
		params.CrontabRecord.Logs = l.logs(params.CrontabRecord.Logs, message)
		functions.LogError("录像计划[4]", message)
		return
	}

	var (
		start = make(map[uint64]*cTypes.ChannelMSRelItem)
		stop  = make(map[uint64]*cTypes.ChannelMSRelItem)
	)
	for _, item := range channelIds {
		start[item] = channelMaps.Data[item]
	}

	for _, item := range stopChannelIds {
		stop[item] = channelMaps.Data[item]
	}

	l.execChannelUniqueIds = channelIds
	if len(start) > 0 {
		l.StartRecordingIds <- start
	}

	if len(stop) > 0 {
		l.StopRecordingIds <- stop
	}
}

func (l *VideoProjectLogic) makeRecord(params *types.CrontabLogicDOParams) {
	for {
		select {
		case v := <-l.StopRecordingIds:
			l.stopRecording(params, v)

		case v := <-l.StartRecordingIds:
			l.startRecording(params, v)
		}
	}
}

func (l *VideoProjectLogic) stopRecording(params *types.CrontabLogicDOParams, maps map[uint64]*cTypes.ChannelMSRelItem) {
	videoProject.NewRecoding().StopRecording(
		&videoProject.Params{
			Mode:          params.SvcCtx.Config.Mode,
			Timeout:       params.CrontabRecord.Timeout,
			GetMSAddress:  func(msIds []uint64) string { return params.SvcCtx.MSVoteNode(msIds).Address },
			VssHttpTarget: params.SvcCtx.Config.VssHttpTarget,
			PlayType:      stream.PlayTypePlay,
			RpcClients:    params.SvcCtx.RpcClients,
		},
		maps,
	)
}

func (l *VideoProjectLogic) startRecording(params *types.CrontabLogicDOParams, maps map[uint64]*cTypes.ChannelMSRelItem) {
	videoProject.NewRecoding().StartRecording(
		&videoProject.Params{
			Mode:          params.SvcCtx.Config.Mode,
			Timeout:       params.CrontabRecord.Timeout,
			GetMSAddress:  func(msIds []uint64) string { return params.SvcCtx.MSVoteNode(msIds).Address },
			VssHttpTarget: params.SvcCtx.Config.VssHttpTarget,
			PlayType:      stream.PlayTypePlay,
			RpcClients:    params.SvcCtx.RpcClients,
		},
		maps,
	)
}
