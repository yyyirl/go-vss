package main

import (
	"flag"
	"os"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/app/sev/db/internal/config"
	"skeyevss/core/app/sev/db/internal/middleware"
	backendservice "skeyevss/core/app/sev/db/internal/server/backendservice"
	configservice "skeyevss/core/app/sev/db/internal/server/configservice"
	deviceservice "skeyevss/core/app/sev/db/internal/server/deviceservice"
	"skeyevss/core/app/sev/db/internal/svc"
	"skeyevss/core/constants"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/pprof"
)

var (
	configFile = flag.String(
		constants.SevParamConfig,
		"etc/.db-rpc.yaml",
		"the config file",
	)

	envFilePath = flag.String(
		constants.SevParamEnv,
		constants.EnvFileNameDev,
		"env file",
	)

	activateCodeFilePath = flag.String(
		constants.SevParamActivateCodePath,
		"etc/"+constants.ActivateCodeFileName,
		"activate code file",
	)

	buildTime = "not set"
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
	c.ActivateCodePath = *activateCodeFilePath

	pprof.Start(c.PProfPort, c.PProfFileDir)
	// if pid := sc.GetPid(c.SevBase.DBPort); pid > 0 {
	// 	if err := sc.KillProcess(pid); err != nil {
	// 		panic(err)
	// 	}
	// }

	// // 验证激活信息
	// if _, err := license.NewVerify().Verification(c.ActivateCodePath); err != nil {
	// 	panic(fmt.Errorf("license 错误: %v", err))
	// }

	functions.PrintStyle("cyan", "------------  [application info]  ------------")
	functions.PrintStyle("red", "[ Application name ]:", c.Name+"...")
	functions.PrintStyle("yellow", "[ Application Config Path ]", *configFile)
	functions.PrintStyle("blue", "[ Application Log Path ]:", c.Log.Path)
	functions.PrintStyle("yellow", "[ Environment ]", c.Mode+"...")
	functions.PrintStyle("red-underline", "[ REDIS ]:", c.RedisHost, "password:", c.CRedis.Pass)

	functions.PrintStyle("red", "[ RPC Service Discovery key ]:", c.Etcd.Key)
	functions.PrintStyle("cyan", "[ Etcd hosts ]:", strings.Join(c.Etcd.Hosts, ",")+"...")
	if c.SevBase.DatabaseType == "mysql" {
		functions.PrintStyle("cyan", "[ DB Base ]:", c.Databases.MysqlBase)
	} else if c.SevBase.DatabaseType == "sqlite" {
		functions.PrintStyle("cyan", "[ DB Base ]:", c.Databases.SqliteBase)
	}
	// functions.PrintStyle("cyan", "mysql host:", c.Databases.Message+"...")
	// functions.PrintStyle("cyan", "redis host:", c.Redis.IP+"...")
	// functions.PrintStyle("cyan", "elasticsearch host:", c.Elasticsearch.IP+"...")
	functions.PrintStyle("red", "Starting [", c.Name, "] rpc server, listen on", c.ListenOn, "...")

	var (
		ctx = svc.NewServiceContext(c)
		sev = zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
			db.RegisterBackendServiceServer(grpcServer, backendservice.NewBackendServiceServer(ctx))
			db.RegisterConfigServiceServer(grpcServer, configservice.NewConfigServiceServer(ctx))
			db.RegisterDeviceServiceServer(grpcServer, deviceservice.NewDeviceServiceServer(ctx))

			if c.Mode == service.DevMode || c.Mode == service.TestMode {
				reflection.Register(grpcServer)
			}
		})
	)
	// rpc中间件拦截器
	sev.AddUnaryInterceptors(middleware.Sev(&c, buildTime))
	defer sev.Stop()

	constants.ENV = c.Mode

	sev.Start()
}
