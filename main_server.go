//go:build server

package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"lanfolder/internal/config"
	"lanfolder/internal/server"
	"lanfolder/internal/share"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	staticFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	cfg := serverConfigFromEnv(config.Load())
	if cfg.Root == "" {
		log.Fatal("shared root is required: set LANFOLDER_ROOT or save a SharedDir in the desktop app")
	}

	srv := server.New(staticFS)
	if err := srv.Start(cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("LanFolder server listening on http://%s:%d\n", cfg.Host, cfg.Port)
	fmt.Printf("Sharing %s with %s permission\n", cfg.Root, cfg.Permission)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Stop(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}

func serverConfigFromEnv(cfg config.Config) server.Config {
	host := firstEnv("LANFOLDER_HOST", "WAILS_SERVER_HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	port := 8080
	if value := firstEnv("LANFOLDER_PORT", "WAILS_SERVER_PORT", "PORT"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			port = parsed
		}
	}

	permission := cfg.Permission
	if value := firstEnv("LANFOLDER_PERMISSION"); value != "" {
		permission = share.Permission(value)
	}
	if !permission.Valid() {
		permission = share.PermissionReadOnly
	}

	return server.Config{
		Host:       host,
		Port:       port,
		Root:       firstNonEmpty(firstEnv("LANFOLDER_ROOT", "SHARED_DIR"), cfg.SharedDir),
		Permission: permission,
		ShowHidden: envBool("LANFOLDER_SHOW_HIDDEN", cfg.ShowHiddenFiles),
	}
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
