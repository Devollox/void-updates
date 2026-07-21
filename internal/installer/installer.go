package installer

import (
	"context"
	"embed"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
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
	wailsruntime.LogInfo(ctx, "Installer.Startup: context set")
}

func embeddedInstallerPrefix() string {
	switch runtime.GOOS {
	case "darwin":
		return "Void.Presence."
	case "linux":
		return "Void.Presence."
	default:
		return "Void.Presence.Setup."
	}
}

func embeddedInstallerExt() string {
	switch runtime.GOOS {
	case "darwin":
		return ".dmg"
	case "linux":
		return ".deb"
	default:
		return ".exe"
	}
}

func (i *Installer) extractInstaller() (string, error) {
	if i.ctx == nil {
		return "", errors.New("no context in extractInstaller")
	}

	wailsruntime.LogInfo(i.ctx, "extractInstaller: start")

	entries, err := i.binaryFolder.ReadDir("binary")
	if err != nil {
		wailsruntime.LogErrorf(i.ctx, "extractInstaller: ReadDir(binary) error: %v", err)
		return "", err
	}

	prefix := embeddedInstallerPrefix()
	ext := embeddedInstallerExt()

	var targetName string
	for _, entry := range entries {
		name := entry.Name()
		wailsruntime.LogInfof(i.ctx, "extractInstaller: found embedded entry: %s (dir=%v)", name, entry.IsDir())
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ext) {
			targetName = name
			break
		}
	}

	if targetName == "" {
		wailsruntime.LogError(i.ctx, "extractInstaller: installer not found in embedded assets")
		return "", errors.New("installer not found in embedded assets (looking for " + prefix + "*" + ext + ")")
	}

	embeddedPath := "binary/" + targetName
	wailsruntime.LogInfof(i.ctx, "extractInstaller: reading embedded file %s", embeddedPath)

	data, err := i.binaryFolder.ReadFile(embeddedPath)
	if err != nil {
		wailsruntime.LogErrorf(i.ctx, "extractInstaller: ReadFile error: %v", err)
		return "", err
	}

	tempDir := os.TempDir()
	extractedPath := filepath.Join(tempDir, targetName)
	wailsruntime.LogInfof(i.ctx, "extractInstaller: writing installer to %s", extractedPath)

	if err := os.WriteFile(extractedPath, data, 0755); err != nil {
		wailsruntime.LogErrorf(i.ctx, "extractInstaller: WriteFile error: %v", err)
		return "", err
	}

	if fi, err := os.Stat(extractedPath); err != nil {
		wailsruntime.LogErrorf(i.ctx, "extractInstaller: Stat after write error: %v", err)
		return "", err
	} else {
		wailsruntime.LogInfof(i.ctx, "extractInstaller: written file size = %d bytes", fi.Size())
	}

	wailsruntime.LogInfo(i.ctx, "extractInstaller: done")
	return extractedPath, nil
}

func killApp() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("taskkill", "/IM", "Void Presence.exe", "/T", "/F")
		hideWindow(cmd)
		_ = cmd.Run()
	case "darwin":
		_ = exec.Command("pkill", "-x", "Void Presence").Run()
	case "linux":
		_ = exec.Command("pkill", "-x", "void-presence").Run()
		_ = exec.Command("pkill", "-x", "Void Presence").Run()
	}
}

func appInstallPath() string {
	switch runtime.GOOS {
	case "darwin":
		return "/Applications/Void Presence.app"
	case "linux":
		return "/usr/bin/void-presence"
	default:
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return ""
		}
		return filepath.Join(localAppData, "Programs", "voidpresence", "Void Presence.exe")
	}
}

func (i *Installer) RunInstalledApp() error {
	if i.ctx == nil {
		return errors.New("no context")
	}

	wailsruntime.LogInfo(i.ctx, "RunInstalledApp: start")

	target := appInstallPath()
	wailsruntime.LogInfof(i.ctx, "RunInstalledApp: target = %s", target)

	if target == "" {
		return errors.New("could not determine app install path")
	}

	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("open", target)
		return cmd.Start()
	case "linux":
		if _, err := os.Stat(target); err != nil {
			wailsruntime.LogErrorf(i.ctx, "RunInstalledApp: Stat error: %v", err)
			return err
		}
		cmd := exec.Command(target)
		return cmd.Start()
	default:
		if _, err := os.Stat(target); err != nil {
			wailsruntime.LogErrorf(i.ctx, "RunInstalledApp: Stat error: %v", err)
			return err
		}
		cmd := exec.Command(target)
		cmd.Dir = filepath.Dir(target)
		if err := cmd.Start(); err != nil {
			wailsruntime.LogErrorf(i.ctx, "RunInstalledApp: cmd.Start error: %v", err)
			return err
		}
	}

	wailsruntime.LogInfo(i.ctx, "RunInstalledApp: started successfully")
	return nil
}

func (i *Installer) cleanupAfterInstall(tempFile string) {
	if i.ctx == nil {
		return
	}

	wailsruntime.LogInfo(i.ctx, "cleanupAfterInstall: start")

	if tempFile != "" {
		if err := os.Remove(tempFile); err != nil && !os.IsNotExist(err) {
			wailsruntime.LogErrorf(i.ctx, "cleanupAfterInstall: Remove(tempFile) error: %v", err)
		} else {
			wailsruntime.LogInfof(i.ctx, "cleanupAfterInstall: removed temp file %s", tempFile)
		}
	}

	cleanupWindowsLeftovers(i.ctx)

	wailsruntime.LogInfo(i.ctx, "cleanupAfterInstall: done")
}

func runPlatformInstaller(path string) (cmd *exec.Cmd, synchronous bool, err error) {
	switch runtime.GOOS {
	case "windows":
		c := exec.Command(path, "/S")
		hideWindow(c)
		return c, false, nil

	case "darwin":
		mountPoint := filepath.Join(os.TempDir(), "VoidPresenceUpdateDMG")
		_ = os.MkdirAll(mountPoint, 0755)
		attach := exec.Command("hdiutil", "attach", "-mountpoint", mountPoint, "-nobrowse", "-quiet", path)
		if e := attach.Run(); e != nil {
			return nil, false, errors.New("hdiutil attach: " + e.Error())
		}
		entries, _ := os.ReadDir(mountPoint)
		for _, e := range entries {
			if filepath.Ext(e.Name()) == ".app" {
				src := filepath.Join(mountPoint, e.Name())
				dst := filepath.Join("/Applications", e.Name())
				if e2 := exec.Command("ditto", src, dst).Run(); e2 != nil {
					_ = exec.Command("hdiutil", "detach", mountPoint, "-quiet").Run()
					return nil, false, errors.New("ditto: " + e2.Error())
				}
				break
			}
		}
		_ = exec.Command("hdiutil", "detach", mountPoint, "-quiet").Run()
		return nil, true, nil

	case "linux":
		if e := exec.Command("pkexec", "dpkg", "-i", path).Run(); e != nil {
			if e2 := exec.Command("sudo", "dpkg", "-i", path).Run(); e2 != nil {
				return nil, false, errors.New("dpkg install failed: " + e2.Error())
			}
		}
		return nil, true, nil

	default:
		return nil, false, errors.New("unsupported platform: " + runtime.GOOS)
	}
}

func (i *Installer) RunBundledInstaller() error {
	if i.ctx == nil {
		return errors.New("no context")
	}

	wailsruntime.LogInfo(i.ctx, "RunBundledInstaller: called")

	if i.started {
		wailsruntime.LogInfo(i.ctx, "RunBundledInstaller: already started, skipping")
		return nil
	}
	i.started = true

	wailsruntime.EventsEmit(i.ctx, "install:progressText", "Launching installer...")
	wailsruntime.LogInfo(i.ctx, "RunBundledInstaller: Launching installer...")

	installerPath, err := i.extractInstaller()
	if err != nil {
		wailsruntime.LogErrorf(i.ctx, "RunBundledInstaller: extractInstaller error: %v", err)
		wailsruntime.EventsEmit(i.ctx, "install:progressText", "Installer extraction failed")
		return err
	}

	wailsruntime.LogInfof(i.ctx, "RunBundledInstaller: installerPath = %s", installerPath)

	killApp()

	cmd, synchronous, err := runPlatformInstaller(installerPath)
	if err != nil {
		wailsruntime.LogErrorf(i.ctx, "RunBundledInstaller: runPlatformInstaller error: %v", err)
		wailsruntime.EventsEmit(i.ctx, "install:progressText", "Failed to launch installer")
		_ = os.Remove(installerPath)
		return err
	}

	if synchronous {
		wailsruntime.LogInfo(i.ctx, "RunBundledInstaller: synchronous install finished")
		wailsruntime.EventsEmit(i.ctx, "install:progressText", "Installer finished")
		i.cleanupAfterInstall(installerPath)
		_ = i.RunInstalledApp()
		time.Sleep(1 * time.Second)
		wailsruntime.Quit(i.ctx)
		return nil
	}

	if err := cmd.Start(); err != nil {
		wailsruntime.LogErrorf(i.ctx, "RunBundledInstaller: cmd.Start error: %v", err)
		wailsruntime.EventsEmit(i.ctx, "install:progressText", "Failed to launch installer")
		_ = os.Remove(installerPath)
		return err
	}

	pid := cmd.Process.Pid
	wailsruntime.LogInfof(i.ctx, "RunBundledInstaller: installer started with PID %d", pid)
	wailsruntime.EventsEmit(i.ctx, "install:progressText", "Installer started")

	go func(pid int, tempFile string) {
		for {
			if !isProcessAlive(pid) {
				wailsruntime.LogInfo(i.ctx, "RunBundledInstaller: installer finished")
				wailsruntime.EventsEmit(i.ctx, "install:progressText", "Installer finished")
				i.cleanupAfterInstall(tempFile)
				_ = i.RunInstalledApp()
				time.Sleep(1 * time.Second)
				wailsruntime.Quit(i.ctx)
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}(pid, installerPath)

	return nil
}
