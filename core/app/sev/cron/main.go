// @Title        cron
// @Description  main
// @Create       yirl 2025/3/13 11:49

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/cron/internal/config"
	"skeyevss/core/app/sev/cron/internal/handler"
	"skeyevss/core/app/sev/cron/internal/logic/crontab"
	fetchdata "skeyevss/core/app/sev/cron/internal/logic/fetch-data"
	"skeyevss/core/app/sev/cron/internal/svc"
	"skeyevss/core/app/sev/cron/internal/types"
	cTypes "skeyevss/core/common/types"
	"skeyevss/core/constants"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/pprof"
)

var (
	buildTime = "not set"

	envFilePath = flag.String(
		constants.SevParamEnv,
		constants.EnvFileNameDev,
		"env file",
	)

	configFile = flag.String(
		constants.SevParamConfig,
		"etc/.cron.yaml",
		"the config file",
	)
)

func main() {
	functions.PrintStyle("yellow", "构建时间: ", strings.ReplaceAll(buildTime, "#", " "))
	functions.PrintStyle("blue", "PARAMS: -"+constants.SevParamEnv, "["+*envFilePath+"] ==> env file")
	functions.PrintStyle("blue", "PARAMS: -"+constants.SevParamConfig, "["+*configFile+"] ==> the config file")
	flag.Parse()

	// 加载环境变量
	functions.OverloadEnvFile(*envFilePath)

	var tz = functions.GetEnvDefault("SKEYEVSS_TZ", "Asia/Shanghai")
	_ = os.Setenv("TZ", tz)
	if _, err := time.LoadLocation(tz); err != nil {
		panic(err)
	}

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	logx.DisableStat()
	logx.MustSetup(c.Log)

	pprof.Start(c.PProfPort, c.PProfFileDir)

	// if pid := sc.GetPid(c.SevBase.CronPort); pid > 0 {
	// 	if err := sc.KillProcess(pid); err != nil {
	// 		panic(err)
	// 	}
	// }

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", c.SevBase.CronPort), nil); err != nil {
			panic(err)
		}
	}()

	var (
		svcCtx = svc.NewServiceContext(c)
		stop   = make(chan os.Signal, 1)
	)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// 基础数据获取
	var fetchRecords = []types.FetchDataProcLogic{
		new(fetchdata.FetchSettingLogic),
		new(fetchdata.FetchCrontabLogic),
		new(fetchdata.FetchMediaServerLogic),
	}
	svcCtx.Data.InitFetchDataState.Add(len(fetchRecords))
	go handler.NewFetchDataLogic(svcCtx).DO(fetchRecords...)
	svcCtx.Data.InitFetchDataState.Wait()
	functions.RestyDebug = c.UseSipPrintLog

	// 消息队列
	// handler.NewQueueHandler(svcCtx).Register(
	// 	new(queue.StreamKeepaliveLogic),
	// )

	// 任务
	handler.NewCrontabHandler(svcCtx).Register(
		&crontab.VideoProjectLogic{
			StartRecordingIds: make(chan map[uint64]*cTypes.ChannelMSRelItem, 10),
			StopRecordingIds:  make(chan map[uint64]*cTypes.ChannelMSRelItem, 10),
		},
	)

	<-stop
}
