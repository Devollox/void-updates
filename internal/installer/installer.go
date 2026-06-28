package installer

import (
	"context"
	"embed"
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
	ctx          context.Context
	started      bool
	binaryFolder embed.FS
}

func NewInstaller(binaryFolder embed.FS) *Installer {
	return &Installer{
		binaryFolder: binaryFolder,
	}
}

func (i *Installer) Startup(ctx context.Context) {
	i.ctx = ctx
	runtime.LogInfo(ctx, "Installer.Startup: context set")
}

func (i *Installer) extractInstaller() (string, error) {
	if i.ctx == nil {
		return "", errors.New("no context in extractInstaller")
	}

	runtime.LogInfo(i.ctx, "extractInstaller: start")

	entries, err := i.binaryFolder.ReadDir("binary")
	if err != nil {
		runtime.LogErrorf(i.ctx, "extractInstaller: ReadDir(binary) error: %v", err)
		return "", err
	}

	var targetName string
	for _, entry := range entries {
		name := entry.Name()
		runtime.LogInfof(i.ctx, "extractInstaller: found embedded entry: %s (dir=%v)", name, entry.IsDir())
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(name, "Void.Presence.Setup.") && strings.HasSuffix(name, ".exe") {
			targetName = name
			break
		}
	}

	if targetName == "" {
		runtime.LogError(i.ctx, "extractInstaller: installer not found in embedded assets")
		return "", errors.New("installer not found in embedded assets")
	}

	embeddedPath := "binary/" + targetName
	runtime.LogInfof(i.ctx, "extractInstaller: reading embedded file %s", embeddedPath)

	data, err := i.binaryFolder.ReadFile(embeddedPath)
	if err != nil {
		runtime.LogErrorf(i.ctx, "extractInstaller: ReadFile error: %v", err)
		return "", err
	}

	tempDir := os.TempDir()
	runtime.LogInfof(i.ctx, "extractInstaller: os.TempDir() = %s", tempDir)

	extractedPath := filepath.Join(tempDir, targetName)
	runtime.LogInfof(i.ctx, "extractInstaller: writing installer to %s", extractedPath)

	if err := os.WriteFile(extractedPath, data, 0755); err != nil {
		runtime.LogErrorf(i.ctx, "extractInstaller: WriteFile error: %v", err)
		return "", err
	}

	if fi, err := os.Stat(extractedPath); err != nil {
		runtime.LogErrorf(i.ctx, "extractInstaller: Stat after write error: %v", err)
		return "", err
	} else {
		runtime.LogInfof(i.ctx, "extractInstaller: written file size = %d bytes", fi.Size())
	}

	runtime.LogInfo(i.ctx, "extractInstaller: done")
	return extractedPath, nil
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

	installerPath, err := i.extractInstaller()
	if err != nil {
		runtime.LogErrorf(i.ctx, "RunBundledInstaller: extractInstaller error: %v", err)
		runtime.EventsEmit(i.ctx, "install:progressText", "Installer extraction failed")
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
		_ = os.Remove(installerPath)
		return err
	}

	pid := cmd.Process.Pid
	runtime.LogInfof(i.ctx, "RunBundledInstaller: installer started with PID %d", pid)

	runtime.EventsEmit(i.ctx, "install:progressText", "Installer started")

	go func(pid int, tempFile string) {
		for {
			if !isProcessAlive(pid) {
				runtime.LogInfo(i.ctx, "RunBundledInstaller: installer finished")
				runtime.EventsEmit(i.ctx, "install:progressText", "Installer finished")
				_ = i.RunInstalledApp()
				_ = os.Remove(tempFile)
				runtime.LogInfof(i.ctx, "RunBundledInstaller: removed temp file %s", tempFile)
				runtime.Quit(i.ctx)
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}(pid, installerPath)

	return nil
}
