// @Title        gbs
// @Description  main
// @Create       yirl 2025/3/13 11:48

package main

import (
	"flag"
	"os"
	"os/signal"
	"skeyevss/core/tps"
	"strings"
	"sync"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/vss/internal/config"
	"skeyevss/core/app/sev/vss/internal/handler/gbs_sip"
	"skeyevss/core/app/sev/vss/internal/logic/gbs_proc"
	"skeyevss/core/app/sev/vss/internal/logic/initialize"
	"skeyevss/core/app/sev/vss/internal/logic/proc"
	"skeyevss/core/app/sev/vss/internal/server"
	"skeyevss/core/app/sev/vss/internal/svc"
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
		"etc/.vss.yaml",
		"the config file",
	)
)

func main() {
	{
		functions.PrintStyle("yellow", "构建时间: ", strings.ReplaceAll(buildTime, "#", " "))
		functions.PrintStyle("blue", "PARAMS: -"+constants.SevParamEnv, "["+*envFilePath+"] ==> env file")
		functions.PrintStyle("blue", "PARAMS: -"+constants.SevParamConfig, "["+*configFile+"] ==> the config file")
	}

	// 解析程序参数
	flag.Parse()
	// 加载env
	functions.OverloadEnvFile(*envFilePath)
	{
		// 设置时区
		var tz = functions.GetEnvDefault("SKEYEVSS_TZ", "Asia/Shanghai")
		_ = os.Setenv("TZ", tz)
		if _, err := time.LoadLocation(tz); err != nil {
			panic(err)
		}
	}

	// 解析配置
	var c config.Config
	{
		conf.MustLoad(*configFile, &c, conf.UseEnv())
		logx.DisableStat()
		logx.MustSetup(c.Log)
		functions.RestyDebug = c.UseSipPrintLog
	}
	// 兼容基础配置
	var baseConf tps.YamlBaseConfig
	conf.MustLoad(*configFile, &baseConf, conf.UseEnv())

	// 性能分析
	pprof.Start(c.PProfPort, c.PProfFileDir)

	var (
		stop   = make(chan os.Signal, 1)
		svcCtx = svc.NewServiceContext(c)
	)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	// 初始化
	initialize.DO(svcCtx, &baseConf)

	// SIP 服务器
	{
		var wg sync.WaitGroup
		wg.Add(2)

		// GBS TCP
		go func() {
			defer wg.Done()
			server.NewSipSev(svcCtx).SipGbsServer(server.SipTCP, gbs_sip.RegisterHandlers(svcCtx))
		}()

		// GBS UDP
		go func() {
			defer wg.Done()
			server.NewSipSev(svcCtx).SipGbsServer(server.SipUDP, gbs_sip.RegisterHandlers(svcCtx))
		}()

		// 国标级联 GBC UDP
		// TODO 完整版请联系作者

		// 国标级联 GBC TCP
		// TODO 完整版请联系作者

		wg.Wait()
	}

	// sse服务器
	go server.NewSSESev(svcCtx).Start()
	// websocket服务器
	go server.NewWSSev(svcCtx).Start()
	// http服务器
	go server.NewHttpSev(svcCtx).Start()

	svcCtx.InitFetchDataState.Add(2)
	// 任务
	server.NewSipProc(svcCtx).DO(
		// 数据获取
		new(proc.FetchDataLogic),

		// ---------------------------------------- gbs
		// 请求发送处理
		new(gbs_proc.SendLogic),
		// 定时发送catalog请求
		new(gbs_proc.CatalogLoopLogic),
		// 心跳检测上线下线
		new(gbs_proc.HeartbeatOfflineLogic),
		// gbs 更新上线下线状态队列
		new(gbs_proc.SetDeviceOnlineStateLogic),
		// 检测设备在线状态
		new(gbs_proc.CheckDeviceOnlineStateLogic),
		// sip日志
		new(gbs_proc.SipLogLogic),
		// ---------------------------------------- gbs

		// ---------------------------------------- gbc 国标级联
		// 设备注册
		// 消息发送
		// TODO 完整版请联系作者
		// ---------------------------------------- gbc 国标级联
	)

	<-stop
	{
		(*svcCtx.GBSTCPSev).Shutdown()
		(*svcCtx.GBSUDPSev).Shutdown()
	}
}
