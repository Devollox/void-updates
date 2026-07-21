//go:build windows

package installer

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
}

func isProcessAlive(pid int) bool {
	h, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)

	status, err := windows.WaitForSingleObject(h, 0)
	if err != nil {
		return false
	}

	return status == 0x00000102
}

func cleanupWindowsLeftovers(ctx context.Context) {
	appDataRoaming := os.Getenv("APPDATA")
	if appDataRoaming != "" {
		setupPrefix := "Void.Presence.Setup"
		updatesPrefix := "Void.Presence.Updates"

		entries, err := os.ReadDir(appDataRoaming)
		if err != nil {
			wailsruntime.LogErrorf(ctx, "cleanupWindowsLeftovers: ReadDir(APPDATA) error: %v", err)
		} else {
			for _, entry := range entries {
				name := entry.Name()
				if strings.HasPrefix(name, setupPrefix) || strings.HasPrefix(name, updatesPrefix) {
					p := filepath.Join(appDataRoaming, name)
					wailsruntime.LogInfof(ctx, "cleanupWindowsLeftovers: removing %s", p)
					if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
						wailsruntime.LogErrorf(ctx, "cleanupWindowsLeftovers: RemoveAll(%s) error: %v", p, err)
					}
				}
			}
		}

		updatesDir := filepath.Join(appDataRoaming, "VoidPresence", "Updates")
		if err := os.RemoveAll(updatesDir); err != nil && !os.IsNotExist(err) {
			wailsruntime.LogErrorf(ctx, "cleanupWindowsLeftovers: RemoveAll(updatesDir) error: %v", err)
		}
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		installerCacheDir := filepath.Join(localAppData, "VoidPresence", "InstallerCache")
		if err := os.RemoveAll(installerCacheDir); err != nil && !os.IsNotExist(err) {
			wailsruntime.LogErrorf(ctx, "cleanupWindowsLeftovers: RemoveAll(installerCacheDir) error: %v", err)
		}
	}
}
