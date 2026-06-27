package installer

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

type Installer struct {
	ctx     context.Context
	started bool
}

func NewInstaller() *Installer {
	return &Installer{}
}

func (i *Installer) Startup(ctx context.Context) {
	i.ctx = ctx
}

func (i *Installer) findInstaller() (string, error) {
    exe, err := os.Executable()
    if err != nil {
        return "", err
    }

    appDir := filepath.Dir(exe)


    installerDir := filepath.Join(appDir, "..", "..", "binary")
    installerDir, _ = filepath.Abs(installerDir)

    entries, err := os.ReadDir(installerDir)
    if err != nil {
        return "", err
    }

    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }

        name := entry.Name()
        if strings.HasPrefix(name, "Void.Presence.Setup.") && strings.HasSuffix(name, ".exe") {
            return filepath.Join(installerDir, name), nil
        }
    }

    return "", errors.New("installer not found")
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

	if status == 0x00000102 {
		return true
	}

	return false
}

func (i *Installer) RunInstalledApp() error {
	if i.ctx == nil {
		return errors.New("no context")
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return errors.New("LOCALAPPDATA not set")
	}

	appPath := filepath.Join(localAppData, "Programs", "voidpresence", "Void Presence.exe")

	if _, err := os.Stat(appPath); err != nil {
		return err
	}

	cmd := exec.Command(appPath)
	cmd.Dir = filepath.Dir(appPath)

	if err := cmd.Start(); err != nil {
		return err
	}

	return nil
}

func (i *Installer) RunBundledInstaller() error {
	if i.ctx == nil {
		return errors.New("no context")
	}

	if i.started {
		return nil
	}
	i.started = true

	runtime.EventsEmit(i.ctx, "install:progressText", "Launching installer...")

	installerPath, err := i.findInstaller()
	if err != nil {
		runtime.EventsEmit(i.ctx, "install:progressText", "Installer not found")
		return err
	}

	appDir := filepath.Dir(installerPath)
	cmd := exec.Command(installerPath, "/S")
	cmd.Dir = appDir

	if err := cmd.Start(); err != nil {
		runtime.EventsEmit(i.ctx, "install:progressText", "Failed to launch installer")
		return err
	}

	pid := cmd.Process.Pid

	runtime.EventsEmit(i.ctx, "install:progressText", "Installer started")

	go func(pid int) {
		for {
			if !isProcessAlive(pid) {
				runtime.EventsEmit(i.ctx, "install:progressText", "Installer finished")

				_ = i.RunInstalledApp()
				runtime.Quit(i.ctx)
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}(pid)

	return nil
}
