//go:build !windows

package system

import "syscall"

// getHiddenWindowAttr returns nil on non-Windows platforms
func getHiddenWindowAttr() *syscall.SysProcAttr {
	return nil
}
