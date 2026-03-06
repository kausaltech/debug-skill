//go:build windows

package dap

import "syscall"

// daemonSysProcAttr returns SysProcAttr for detaching the daemon process.
func daemonSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
