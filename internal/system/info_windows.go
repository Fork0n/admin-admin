//go:build windows

package system

import "syscall"

// getHiddenWindowAttr returns syscall attributes to hide command windows on Windows
func getHiddenWindowAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true}
}
