package main

import (
  "context"
  "embed"
  "log"

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

  err := wails.Run(&options.App{
    Title:  "Void Presence Updates",
    Width:  480,
    Height: 148,
    Frameless: true,
		DisableResize: true,
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
