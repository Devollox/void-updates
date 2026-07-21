//go:build !windows

package installer

import (
	"context"
	"os/exec"
	"syscall"
)

func hideWindow(_ *exec.Cmd) {}

func isProcessAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil
}

func cleanupWindowsLeftovers(_ context.Context) {}
