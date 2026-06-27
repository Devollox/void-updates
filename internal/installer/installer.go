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
	runtime.LogInfo(ctx, "Installer.Startup: context set")
}

func (i *Installer) findLocalInstaller() (string, error) {
	if i.ctx == nil {
		return "", errors.New("no context in findLocalInstaller")
	}

	runtime.LogInfo(i.ctx, "findLocalInstaller: start")

	exePath, err := os.Executable()
	if err != nil {
		runtime.LogErrorf(i.ctx, "findLocalInstaller: os.Executable error: %v", err)
		return "", err
	}
	exeDir := filepath.Dir(exePath)
	runtime.LogInfof(i.ctx, "findLocalInstaller: exeDir = %s", exeDir)

	entries, err := os.ReadDir(exeDir)
	if err != nil {
		runtime.LogErrorf(i.ctx, "findLocalInstaller: ReadDir error: %v", err)
		return "", err
	}

	var targetName string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		runtime.LogInfof(i.ctx, "findLocalInstaller: found entry: %s", name)
		if strings.HasPrefix(name, "Void.Presence.Setup.") && strings.HasSuffix(name, ".exe") {
			targetName = name
			break
		}
	}

	if targetName == "" {
		runtime.LogError(i.ctx, "findLocalInstaller: installer not found near updater exe")
		return "", errors.New("installer not found near updater exe")
	}

	fullPath := filepath.Join(exeDir, targetName)
	runtime.LogInfof(i.ctx, "findLocalInstaller: installer path = %s", fullPath)

	if _, err := os.Stat(fullPath); err != nil {
		runtime.LogErrorf(i.ctx, "findLocalInstaller: Stat error: %v", err)
		return "", err
	}

	runtime.LogInfo(i.ctx, "findLocalInstaller: done")
	return fullPath, nil
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

	runtime.LogInfo(i.ctx, "RunInstalledApp: start")

	localAppData := os.Getenv("LOCALAPPDATA")
	runtime.LogInfof(i.ctx, "RunInstalledApp: LOCALAPPDATA = %s", localAppData)

	if localAppData == "" {
		return errors.New("LOCALAPPDATA not set")
	}

	appPath := filepath.Join(localAppData, "Programs", "voidpresence", "Void Presence.exe")
	runtime.LogInfof(i.ctx, "RunInstalledApp: appPath = %s", appPath)

	if _, err := os.Stat(appPath); err != nil {
		runtime.LogErrorf(i.ctx, "RunInstalledApp: Stat error: %v", err)
		return err
	}

	cmd := exec.Command(appPath)
	cmd.Dir = filepath.Dir(appPath)
	runtime.LogInfof(i.ctx, "RunInstalledApp: starting %s (Dir=%s)", appPath, cmd.Dir)

	if err := cmd.Start(); err != nil {
		runtime.LogErrorf(i.ctx, "RunInstalledApp: cmd.Start error: %v", err)
		return err
	}

	runtime.LogInfo(i.ctx, "RunInstalledApp: started successfully")
	return nil
}

func (i *Installer) RunBundledInstaller() error {
	if i.ctx == nil {
		return errors.New("no context")
	}

	runtime.LogInfo(i.ctx, "RunBundledInstaller: called")

	if i.started {
		runtime.LogInfo(i.ctx, "RunBundledInstaller: already started, skipping")
		return nil
	}
	i.started = true

	runtime.EventsEmit(i.ctx, "install:progressText", "Launching installer...")
	runtime.LogInfo(i.ctx, "RunBundledInstaller: Launching installer...")

	installerPath, err := i.findLocalInstaller()
	if err != nil {
		runtime.LogErrorf(i.ctx, "RunBundledInstaller: findLocalInstaller error: %v", err)
		runtime.EventsEmit(i.ctx, "install:progressText", "Installer not found")
		return err
	}

	runtime.LogInfof(i.ctx, "RunBundledInstaller: installerPath = %s", installerPath)

	appDir := filepath.Dir(installerPath)
	runtime.LogInfof(i.ctx, "RunBundledInstaller: appDir = %s", appDir)

	cmd := exec.Command(installerPath, "/S")
	cmd.Dir = appDir
	runtime.LogInfof(i.ctx, "RunBundledInstaller: exec.Command(%s, /S), Dir=%s", installerPath, cmd.Dir)

	if err := cmd.Start(); err != nil {
		runtime.LogErrorf(i.ctx, "RunBundledInstaller: cmd.Start error: %v", err)
		runtime.EventsEmit(i.ctx, "install:progressText", "Failed to launch installer")
		return err
	}

	pid := cmd.Process.Pid
	runtime.LogInfof(i.ctx, "RunBundledInstaller: installer started with PID %d", pid)

	runtime.EventsEmit(i.ctx, "install:progressText", "Installer started")

	go func(pid int) {
		for {
			if !isProcessAlive(pid) {
				runtime.LogInfo(i.ctx, "RunBundledInstaller: installer finished")
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
