// @Title        sys
// @Description  main
// @Create       yiyiyi 2025/9/9 14:52

package common

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"skeyevss/core/app/sev/backend/internal/svc"
	"skeyevss/core/constants"
	"skeyevss/core/localization"
	"skeyevss/core/pkg/contextx"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/functions/sc"
	"skeyevss/core/pkg/response"
)

func DepartmentIds(ctx context.Context) []uint64 {
	if contextx.GetSuperState(ctx) > 0 {
		return nil
	}

	var departmentIds = contextx.GetDepartmentIds(ctx)
	if len(departmentIds) > 0 {
		return departmentIds
	}

	return []uint64{}
}

func Restart(svcCtx *svc.ServiceContext, zipFilepath string) *response.HttpErr {
	if svcCtx.Config.OSEnvironment == "docker" {
		return dockerRestart(svcCtx)
	}

	return binRestart(svcCtx, zipFilepath)
}

func dockerRestart(svcCtx *svc.ServiceContext) *response.HttpErr {
	// 下载 启动脚本
	var packagePath = "/app/tmp/latest/docker.image.start.tar.gz"
	if err := functions.DownloadFile(fmt.Sprintf("%s/skeyevss/packages/docker.image.start.tar.gz", svcCtx.Config.MinioApiTarget), packagePath); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00278)
	}

	if err := functions.UnTarGz(packagePath, "/app/tmp/latest"); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00279)
	}

	if err := functions.Mv("/app/tmp/latest/start/start.sh", "/app/tmp/latest/start.sh"); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00279)
	}

	if err := functions.Mv("/app/tmp/latest/start/docker-compose.yml", "/app/tmp/latest/docker-compose.yml"); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00279)
	}

	_ = os.RemoveAll("/app/tmp/latest/start")

	// 执行宿主机脚本
	var (
		scriptPath    = "/app/tmp/restart.sh"
		scriptContent = `#!/bin/bash
nohup ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@` + svcCtx.Config.InternalIP + ` "nohup sh ` + svcCtx.Config.SevVolumesDir + `/start.sh > ` + svcCtx.Config.SevVolumesDir + `/logs/restart.log 2>&1 &" > /app/logs/restart.c.log 2>&1 &
`
	)
	if err := functions.WriteToFile(scriptPath, scriptContent); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00275)
	}

	if _, err := sc.SyscallScriptCrossPlatformDetach(scriptPath); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00276)
	}

	return nil
}

func binRestart(svcCtx *svc.ServiceContext, zipFilepath string) *response.HttpErr {
	if svcCtx.Config.Mode == constants.ENV_DEVELOPMENT {
		return nil
	}

	var (
		scriptPath    = path.Join(svcCtx.Config.SevBase.Root, "update.server")
		guardBinPath  = path.Join(svcCtx.Config.SevRes.AssetDir, "sev", svcCtx.Config.SevBase.SevNameGuard)
		scriptContent string
	)
	// if zipFilepath != "" {
	// 	if !strings.HasPrefix(zipFilepath, svcCtx.Config.SavePath.File) {
	// 		zipFilepath = filepath.ToSlash(path.Join(svcCtx.Config.SevRes.AssetDir, "sev", zipFilepath))
	// 	}
	// }

	if runtime.GOOS == "windows" {
		guardBinPath += ".exe"
	}

	if runtime.GOOS == "windows" {
		scriptPath += ".bat"
		scriptPath = functions.ConvertPath(scriptPath)
		if zipFilepath == "" {
			scriptContent = `@echo off
start /b ` + functions.ConvertPath(guardBinPath) + ` restart`
		} else {
			// if "%1" == "h" goto begin
			// start mshta vbscript:createobject("wscript.shell").run("%~nx0 h",0)(window.close)&&exit
			// :begin
			scriptContent = `@echo off
start /b ` + functions.ConvertPath(guardBinPath) + ` stop
timeout /t 2 /nobreak
if exist "` + functions.ConvertPath(strings.TrimRight(svcCtx.Config.SevBase.Root, "/")+`/`+constants.EnvFileNameOld) + `" (
    del ` + functions.ConvertPath(strings.TrimRight(svcCtx.Config.SevBase.Root, "/")+`/`+constants.EnvFileNameOld) + `
)
copy ` + functions.ConvertPath(strings.TrimRight(svcCtx.Config.SevBase.Root, "/")+`/`+constants.EnvFileNameProd) + ` ` + functions.ConvertPath(strings.TrimRight(svcCtx.Config.SevBase.Root, "/")+`/`+constants.EnvFileNameOld) + `
powershell -Command "Expand-Archive -Path '` + zipFilepath + `' -DestinationPath '` + svcCtx.Config.SevBase.Root + `' -Force"

start /b ` + functions.ConvertPath(guardBinPath) + ` start`
		}
	} else {
		scriptPath += ".sh"
		if zipFilepath == "" {
			scriptContent = `#!/bin/bash
` + guardBinPath + ` restart`
		} else {
			scriptContent = `#!/bin/bash
rm -rf ` + scriptPath + `.log
echo "rm -rf ` + strings.TrimRight(svcCtx.Config.SevBase.Root, "/") + `/` + constants.EnvFileNameOld + `" >> ` + scriptPath + `.log
rm -rf ` + strings.TrimRight(svcCtx.Config.SevBase.Root, "/") + `/` + constants.EnvFileNameOld + `
echo "cp ` + strings.TrimRight(svcCtx.Config.SevBase.Root, "/") + `/` + constants.EnvFileNameProd + `` + strings.TrimRight(svcCtx.Config.SevBase.Root, "/") + `/` + constants.EnvFileNameOld + `" >> ` + scriptPath + `.log
cp ` + strings.TrimRight(svcCtx.Config.SevBase.Root, "/") + `/` + constants.EnvFileNameProd + `` + strings.TrimRight(svcCtx.Config.SevBase.Root, "/") + `/` + constants.EnvFileNameOld + `
echo "unzip -q -o ` + zipFilepath + ` -d ` + svcCtx.Config.SevBase.Root + `" >> ` + scriptPath + `.log
unzip -q -o ` + zipFilepath + ` -d ` + svcCtx.Config.SevBase.Root + `
echo "` + guardBinPath + ` restart" >> ` + scriptPath + `.log
` + guardBinPath + ` restart`
		}
	}

	_ = os.Remove(scriptPath)
	if runtime.GOOS == "windows" {
		if err := functions.WriteToFileWithLineEnding(scriptPath, scriptContent, "crlf"); err != nil {
			return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00275)
		}
	} else {
		if err := functions.WriteToFile(scriptPath, scriptContent); err != nil {
			return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00275)
		}
	}

	if _, err := sc.SyscallScriptCrossPlatformDetach(scriptPath); err != nil {
		return response.MakeError(response.NewHttpRespMessage().Err(err), localization.M00276)
	}

	return nil
}
