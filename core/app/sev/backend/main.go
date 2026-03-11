package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	"skeyevss/core/app/sev/backend/internal/config"
	"skeyevss/core/app/sev/backend/internal/handler"
	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/constants"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/pprof"
)

var (
	configFile = flag.String(
		constants.SevParamConfig,
		"etc/.backend-api.yaml",
		"the config file",
	)

	envFilePath = flag.String(
		constants.SevParamEnv,
		constants.EnvFileNameDev,
		"env file",
	)
	buildTime = "dev-build" // debug: -ldflags="-X main.buildTime=dev-build"
)

func main() {
	if buildTime == "dev-build" {
		buildTime = functions.NewTimer().Format("")
		functions.PrintStyle("yellow", "构建时间: dev-build: ", buildTime)
	} else {
		functions.PrintStyle("yellow", "构建时间: ", strings.ReplaceAll(buildTime, "#", " "))
		if buildTime == "not set" {
			buildTime = ""
		}
	}

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
	c.EnvFile = *envFilePath
	functions.RestyDebug = c.UseSipPrintLog

	pprof.Start(c.PProfPort, c.PProfFileDir)

	// if c.Mode != constants.ENV_DEVELOPMENT && !sc.IsAdmin() {
	// 	functions.PrintStyle("red", "请使用管理员权限运行 .")
	// 	time.Sleep(time.Second)
	// 	os.Exit(1)
	// }

	// if pid := sc.GetPid(c.SevBase.BackendApiPort); pid > 0 {
	// 	if err := sc.KillProcess(pid); err != nil {
	// 		panic(err)
	// 	}
	// }

	constants.ENV = c.Mode

	functions.PrintStyle("cyan", "------------  [application info]  ------------")
	functions.PrintStyle("blue", "[ Application Name ]:", c.Name)
	functions.PrintStyle("red", "[ Listen ]: ", c.Host+":"+strconv.Itoa(c.Port))
	functions.PrintStyle("red", "[ Environment ]: ", c.Mode)
	functions.PrintStyle("blue", "[ Application Config Path ]:", *configFile)
	functions.PrintStyle("blue", "[ Application Log Path ]:", c.Log.Path)
	functions.PrintStyle("red-underline", "[ REDIS ]:", c.RedisHost)
	functions.PrintStyle("green", "Starting Server At", c.Host+":"+strconv.Itoa(c.Port)+"...\n")

	var server = rest.MustNewServer(c.RestConf)
	defer server.Stop()

	var svcCtx = svc.NewServiceContext(c, buildTime)
	svcCtx.BuildTime = buildTime
	handler.RegisterHandlers(server, svcCtx)

	server.Start()
}
