package main

import (
    "context"
    "embed"
    "log"
    "os"
    "path/filepath"
    "strings"

    "local/void-updates/internal/installer"

    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("panic: %v", r)
        }
    }()

    app := NewApp()
    inst := installer.NewInstaller()

    title := buildTitleFromInstaller()

    err := wails.Run(&options.App{
        Title:          title,
        Width:          480,
        Height:         148,
        Frameless:      true,
        DisableResize:  true,
        AssetServer: &assetserver.Options{
            Assets: assets,
        },
        BackgroundColour: &options.RGBA{R: 5, G: 5, B: 5, A: 1},
        OnStartup: func(ctx context.Context) {
            app.startup(ctx)
            inst.Startup(ctx)
        },
        Bind: []interface{}{
            app,
            inst,
        },
    })

    if err != nil {
        log.Printf("wails error: %v", err)
    }
}

func buildTitleFromInstaller() string {
    exe, err := os.Executable()
    if err != nil {
        return "Void Presence Updates"
    }

    appDir := filepath.Dir(exe)

    installerDir := filepath.Join(appDir, "..", "..", "binary")
    installerDir, _ = filepath.Abs(installerDir)

    entries, err := os.ReadDir(installerDir)
    if err != nil {
        return "Void Presence Updates"
    }

    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        name := entry.Name()
        if strings.HasPrefix(name, "Void.Presence.Setup.") &&
            strings.HasSuffix(name, ".exe") {

            base := strings.TrimSuffix(name, ".exe")
            parts := strings.Split(base, ".")
            if len(parts) >= 4 {
                version := strings.Join(parts[3:], ".")
                return "Void Presence Updates " + version
            }
        }
    }

    return "Void Presence Updates"
}
