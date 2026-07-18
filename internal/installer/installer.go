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
  "syscall"

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

func (i *Installer) cleanupAfterInstall(tempFile string) {
  if i.ctx == nil {
    return
  }

  runtime.LogInfo(i.ctx, "cleanupAfterInstall: start")

  if tempFile != "" {
    if err := os.Remove(tempFile); err != nil && !os.IsNotExist(err) {
      runtime.LogErrorf(i.ctx, "cleanupAfterInstall: Remove(tempFile) error: %v", err)
    } else {
      runtime.LogInfof(i.ctx, "cleanupAfterInstall: removed temp file %s", tempFile)
    }
  }

  appDataRoaming := os.Getenv("APPDATA")
  if appDataRoaming != "" {
    setupPrefix := "Void.Presence.Setup"
    updatesPrefix := "Void.Presence.Updates"

    entries, err := os.ReadDir(appDataRoaming)
    if err != nil {
      runtime.LogErrorf(i.ctx, "cleanupAfterInstall: ReadDir(APPDATA) error: %v", err)
    } else {
      for _, entry := range entries {
        if !entry.IsDir() {
          continue
        }
        name := entry.Name()
        if strings.HasPrefix(name, setupPrefix) && strings.HasSuffix(name, ".exe") {
          dirPath := filepath.Join(appDataRoaming, name)
          runtime.LogInfof(i.ctx, "cleanupAfterInstall: removing setup dir %s", dirPath)
          if err := os.RemoveAll(dirPath); err != nil && !os.IsNotExist(err) {
            runtime.LogErrorf(i.ctx, "cleanupAfterInstall: RemoveAll(setup dir) error: %v", err)
          }
        }
        if strings.HasPrefix(name, updatesPrefix) && strings.HasSuffix(name, ".exe") {
          dirPath := filepath.Join(appDataRoaming, name)
          runtime.LogInfof(i.ctx, "cleanupAfterInstall: removing updates dir %s", dirPath)
          if err := os.RemoveAll(dirPath); err != nil && !os.IsNotExist(err) {
            runtime.LogErrorf(i.ctx, "cleanupAfterInstall: RemoveAll(updates dir) error: %v", err)
          }
        }
      }
    }

    updatesDir := filepath.Join(appDataRoaming, "VoidPresence", "Updates")
    runtime.LogInfof(i.ctx, "cleanupAfterInstall: updatesDir = %s", updatesDir)

    if err := os.RemoveAll(updatesDir); err != nil && !os.IsNotExist(err) {
      runtime.LogErrorf(i.ctx, "cleanupAfterInstall: RemoveAll(updatesDir) error: %v", err)
    } else {
      runtime.LogInfo(i.ctx, "cleanupAfterInstall: updatesDir removed or did not exist")
    }
  } else {
    runtime.LogInfo(i.ctx, "cleanupAfterInstall: APPDATA not set, skipping roaming cleanup")
  }

  localAppData := os.Getenv("LOCALAPPDATA")
  if localAppData != "" {
    installerCacheDir := filepath.Join(localAppData, "VoidPresence", "InstallerCache")
    runtime.LogInfof(i.ctx, "cleanupAfterInstall: installerCacheDir = %s", installerCacheDir)

    if err := os.RemoveAll(installerCacheDir); err != nil && !os.IsNotExist(err) {
      runtime.LogErrorf(i.ctx, "cleanupAfterInstall: RemoveAll(installerCacheDir) error: %v", err)
    } else {
      runtime.LogInfo(i.ctx, "cleanupAfterInstall: installerCacheDir removed or did not exist")
    }
  } else {
    runtime.LogInfo(i.ctx, "cleanupAfterInstall: LOCALAPPDATA not set, skipping local cleanup")
  }

  runtime.LogInfo(i.ctx, "cleanupAfterInstall: done")
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

  kill := exec.Command("taskkill", "/IM", "Void Presence.exe", "/T", "/F")
  kill.SysProcAttr = &syscall.SysProcAttr{
    HideWindow:    true,
    CreationFlags: 0x08000000,
  }
  _ = kill.Run()

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
        i.cleanupAfterInstall(tempFile)
        _ = i.RunInstalledApp()
        runtime.Quit(i.ctx)
        return
      }
      time.Sleep(500 * time.Millisecond)
    }
  }(pid, installerPath)

  return nil
}
