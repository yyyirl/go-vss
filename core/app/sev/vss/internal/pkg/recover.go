package pkg

import (
	"fmt"
	"runtime/debug"
	"time"

	"skeyevss/core/app/sev/vss/internal/config"
	"skeyevss/core/pkg/functions"
)

func NewRecover(_ *config.Config, call func()) {
	if err := recover(); err != nil {
		var (
			broken = string(debug.Stack())
			info   = fmt.Sprintf("error:%+v\nbroken: %s", err, broken)
		)

		functions.LogError("recover info:", info)

		// 防止高频panic
		time.Sleep(1 * time.Second)
		if call != nil {
			functions.LogInfo("restart interval server ...")
			call()
		}
	}
}
