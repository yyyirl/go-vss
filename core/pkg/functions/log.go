/**
 * @Author:         yi
 * @Description:    log
 * @Version:        1.0.0
 * @Date:           2022/10/11 11:14
 */
package functions

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// var mainIp = GetEnvDefault("MAIN_IP", "localhost")

func LogAlert(v string) {
	logx.Alert(v + Caller(2))
}

func LogError(v ...interface{}) {
	// v = append(v, " main-ip: "+mainIp)
	// logx.Error(append(v, Caller(2))...)
	logx.WithCallerSkip(1).Error(v...)
}

func LogcError(ctx context.Context, v ...interface{}) {
	// v = append(v, " main-ip: "+mainIp)
	// logx.Error(append(v, Caller(2))...)
	logx.WithCallerSkip(1).WithContext(ctx).Error(v...)
}

func LogInfo(v ...interface{}) {
	// v = append(v, " main-ip: "+mainIp)
	// logx.Info(append(v, Caller(2))...)
	logx.WithCallerSkip(1).Info(v...)
}

func LogcInfo(ctx context.Context, v ...interface{}) {
	// v = append(v, " main-ip: "+mainIp)
	// logx.Info(append(v, Caller(2))...)
	logx.WithCallerSkip(1).WithContext(ctx).Info(v...)
}
