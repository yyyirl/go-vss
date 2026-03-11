//go:build !windows
// +build !windows

package sc

import "syscall"

func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
		Setsid:  true,
	}
}
