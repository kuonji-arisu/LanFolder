//go:build !server

package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"lanfolder/internal/config"
	"lanfolder/internal/desktop"
	"lanfolder/internal/platform"
	"lanfolder/internal/server"
	"lanfolder/internal/share"
)

type AppService struct {
	mu     sync.Mutex
	app    *application.App
	window application.Window
	server *server.Server
	config config.Config
}

func (s *AppService) State() desktop.AppState {
	s.mu.Lock()
	cfg := s.config
	s.mu.Unlock()
	return s.snapshot(cfg)
}

func (s *AppService) ChooseFolder() (desktop.AppState, error) {
	result, err := s.app.Dialog.OpenFile().
		CanChooseDirectories(true).
		CanCreateDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()
	if err != nil {
		return s.State(), err
	}
	if result == "" {
		return s.State(), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	nextConfig := s.config
	nextConfig.SharedDir = result
	if err := config.Save(nextConfig); err != nil {
		return s.snapshot(s.config), err
	}
	s.config = nextConfig
	if s.server.Status().Running {
		if err := s.restartLocked(); err != nil {
			return s.snapshot(s.config), err
		}
	}
	return s.snapshot(s.config), nil
}

func (s *AppService) SaveSettings(cfg config.Config) (desktop.AppState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return s.snapshot(s.config), fmt.Errorf("端口必须在 1 到 65535 之间")
	}
	if !cfg.Permission.Valid() {
		cfg.Permission = share.PermissionReadOnly
	}
	if cfg.StartAtLogin != s.config.StartAtLogin {
		if err := platform.SetStartAtLogin(cfg.StartAtLogin); err != nil {
			return s.snapshot(s.config), err
		}
	}
	wasRunning := s.server.Status().Running
	shouldRestart := wasRunning && serverConfigChanged(s.config, cfg)
	if err := config.Save(cfg); err != nil {
		return s.snapshot(s.config), err
	}
	s.config = cfg
	if shouldRestart {
		if err := s.restartLocked(); err != nil {
			return s.snapshot(s.config), err
		}
	}
	return s.snapshot(s.config), nil
}

func (s *AppService) StartSharing() (desktop.AppState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.config.SharedDir == "" {
		return s.snapshot(s.config), fmt.Errorf("请先选择共享目录")
	}
	err := s.startSharingLocked()
	return s.snapshot(s.config), err
}

func (s *AppService) StopSharing() (desktop.AppState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := s.server.Stop(ctx)
	return s.snapshot(s.config), err
}

func (s *AppService) startSharingLocked() error {
	return s.server.Start(server.Config{
		Host:       "0.0.0.0",
		Port:       s.config.Port,
		Root:       s.config.SharedDir,
		Permission: s.config.Permission,
		ShowHidden: s.config.ShowHiddenFiles,
	})
}

func (s *AppService) OpenSharedFolder() error {
	s.mu.Lock()
	sharedDir := s.config.SharedDir
	s.mu.Unlock()
	if sharedDir == "" {
		return fmt.Errorf("请先选择共享目录")
	}
	return platform.OpenFolder(sharedDir)
}

func (s *AppService) Logs() []server.LogEntry {
	return s.server.Logs()
}

func (s *AppService) restartLocked() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.server.Stop(ctx); err != nil {
		return err
	}
	return s.startSharingLocked()
}

func (s *AppService) snapshot(cfg config.Config) desktop.AppState {
	return desktop.SnapshotBuilder{App: s.app, Window: s.window}.Build(cfg, s.server.Status(), s.addresses(cfg))
}

func (s *AppService) addresses(cfg config.Config) []string {
	ips := platform.LANIPs()
	if len(ips) == 0 {
		ips = []string{"127.0.0.1"}
	}
	addrs := make([]string, 0, len(ips))
	for _, ip := range ips {
		addrs = append(addrs, fmt.Sprintf("http://%s:%d", ip, cfg.Port))
	}
	return addrs
}

func (s *AppService) emitStateChanged(reason string) {
	if s.app != nil {
		s.app.Event.Emit("app:state-changed", map[string]string{"reason": reason})
	}
}

func serverConfigChanged(previous config.Config, next config.Config) bool {
	return previous.SharedDir != next.SharedDir ||
		previous.Port != next.Port ||
		previous.Permission != next.Permission ||
		previous.ShowHiddenFiles != next.ShowHiddenFiles
}
