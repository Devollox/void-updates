package main

import (
	"context"
	"embed"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"local/void-updates/internal/installer"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed binary/*
var binaryFolder embed.FS

func logDir() string {
	switch runtime.GOOS {
	case "windows":
		if ld := os.Getenv("LOCALAPPDATA"); ld != "" {
			return filepath.Join(ld, "voidupdates")
		}
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Logs", "voidupdates")
		}
	case "linux":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".local", "share", "voidupdates")
		}
	}
	return os.TempDir()
}

func initLogging() {
	exe, _ := os.Executable()

	dir := logDir()
	_ = os.MkdirAll(dir, 0755)

	logPath := filepath.Join(dir, "updater.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	log.Printf("=== Void Presence Updates started ===")
	log.Printf("exe = %s", exe)
	log.Printf("log file = %s", logPath)
}

func main() {
	initLogging()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic: %v", r)
		}
	}()

	log.Printf("main: creating app & installer")

	app := NewApp()
	inst := installer.NewInstaller(binaryFolder)
	title := "Void Presence Updates"

	log.Printf("main: starting Wails app with title %q", title)

	err := wails.Run(&options.App{
		Title:         title,
		Width:         480,
		Height:        148,
		Frameless:     true,
		DisableResize: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 5, G: 5, B: 5, A: 1},
		OnStartup: func(ctx context.Context) {
			log.Printf("main: OnStartup called")
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

	log.Printf("main: exiting")
}
