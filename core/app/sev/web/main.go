// @Title        web服务器
// @Description  main
// @Create       yirl 2025/3/13 11:47

package main

import (
	"flag"
	"os"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/app/sev/web/internal/config"
	"skeyevss/core/app/sev/web/internal/svc"
	"skeyevss/core/constants"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/pprof"
)

var (
	buildTime = "not set"

	webStaticDir = flag.String(
		constants.SevWebParamWebStaticDir,
		"",
		"网站管理后台前端代码路径",
	)
	certPemPath = flag.String(
		constants.SevWebParamCertPem,
		"",
		"证书路径.pem",
	)
	certKeyPath = flag.String(
		constants.SevWebParamCertKey,
		"",
		"证书路径.key",
	)

	envFilePath = flag.String(
		constants.SevParamEnv,
		constants.EnvFileNameDev,
		"env file",
	)

	configFile = flag.String(
		constants.SevParamConfig,
		"etc/.web-sev.yaml",
		"the config file",
	)
)

func main() {
	flag.Parse()

	functions.PrintStyle("yellow", "构建时间: ", strings.ReplaceAll(buildTime, "#", " "))
	functions.PrintStyle("blue", "PARAMS: -"+constants.SevWebParamWebStaticDir, "["+*webStaticDir+"] ==> 网站管理后台前端代码路径")
	functions.PrintStyle("blue", "PARAMS: -"+constants.SevParamEnv, "["+*envFilePath+"] ==> env file")
	functions.PrintStyle("blue", "PARAMS: -"+constants.SevParamConfig, "["+*configFile+"] ==> the config file")
	if webStaticDir == nil {
		panic("web static dir is nil")
	}

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

	var (
		realCertPemPath = *webStaticDir
		realCertKeyPath = *certPemPath
	)
	if realCertPemPath == "" || realCertKeyPath == "" {
		realCertPemPath = c.SevBase.SSL.PublicKey
		realCertKeyPath = c.SevBase.SSL.PrivateKey
	}

	functions.PrintStyle("blue", "PARAMS: -"+constants.SevWebParamCertPem, "["+realCertPemPath+"] ==> 证书路径.pem")
	functions.PrintStyle("blue", "PARAMS: -"+constants.SevWebParamCertKey, "["+realCertKeyPath+"] ==> 证书路径.key")

	// if pid := sc.GetPid(c.SevBase.WebSevPort); pid > 0 {
	// 	if err := sc.KillProcess(pid); err != nil {
	// 		panic(err)
	// 	}
	// }

	functions.PrintStyle("cyan", "------------  [application info]  ------------")
	functions.PrintStyle("red", "[ Listen ]: ", c.SevBase.WebSevPort)
	functions.PrintStyle("red", "[ Environment ]: ", c.Mode)
	functions.PrintStyle("blue", "[ Application Name ]:", c.Name)
	functions.PrintStyle("blue", "[ Application Log Path ]:", c.Log.Path)

	svc.NewProxy(&c, *webStaticDir, *certPemPath, *certKeyPath).Start()
}
