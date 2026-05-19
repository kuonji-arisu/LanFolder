package appservice

import (
	"context"
	"time"

	"lanfolder/internal/config"
	"lanfolder/internal/desktop"
	"lanfolder/internal/platform"
	"lanfolder/internal/server"
)

type ShareService struct {
	server *server.Server
}

func NewShareService(server *server.Server) *ShareService {
	return &ShareService{server: server}
}

func (s *ShareService) Start(cfg config.Config) error {
	return s.server.Start(s.ServerConfig(cfg))
}

func (s *ShareService) Stop(ctx context.Context) error {
	return s.server.Stop(ctx)
}

func (s *ShareService) Restart(cfg config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.Stop(ctx); err != nil {
		return err
	}
	return s.Start(cfg)
}

func (s *ShareService) ServerConfig(cfg config.Config) server.Config {
	return server.Config{
		Host:                  "0.0.0.0",
		Port:                  cfg.Port,
		Root:                  cfg.SharedDir,
		Permission:            cfg.Permission,
		ShowHidden:            cfg.ShowHiddenFiles,
		AccessApproval:        cfg.AccessApproval,
		AccessSessionLifetime: cfg.AccessSessionLifetime,
		Language:              cfg.Language,
	}
}

func (s *AppService) StartSharing() (desktop.AppState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.config.SharedDir == "" {
		return s.snapshot(s.config), newCommandError(errSharedDirRequired, nil)
	}
	err := s.startSharingLocked()
	return s.snapshot(s.config), err
}

func (s *AppService) autoStartSharing() {
	s.mu.Lock()
	autoShare := s.config.AutoShare
	accessApproval := s.config.AccessApproval
	sharedDir := s.config.SharedDir
	s.mu.Unlock()
	if !autoShare {
		return
	}
	if !accessApproval {
		s.addNotice(desktop.NoticeError, desktop.NoticeSourceStartup, newCommandError(errAccessApprovalRequired, nil), "")
		return
	}
	if sharedDir == "" {
		s.addNotice(desktop.NoticeError, desktop.NoticeSourceStartup, newCommandError(errSharedDirRequired, nil), "")
		return
	}
	if _, err := s.StartSharing(); err != nil {
		s.addNotice(desktop.NoticeError, desktop.NoticeSourceStartup, err, "")
	}
}

func (s *AppService) StopSharing() (desktop.AppState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := s.shareServiceLocked().Stop(ctx)
	return s.snapshot(s.config), err
}

func (s *AppService) startSharingLocked() error {
	return s.shareServiceLocked().Start(s.config)
}

func (s *AppService) OpenSharedFolder() error {
	s.mu.Lock()
	sharedDir := s.config.SharedDir
	s.mu.Unlock()
	if sharedDir == "" {
		return newCommandError(errSharedDirRequired, nil)
	}
	return platform.OpenFolder(sharedDir)
}

func (s *AppService) restartLocked() error {
	return s.shareServiceLocked().Restart(s.config)
}

func serverConfigChanged(previous config.Config, next config.Config) bool {
	return SettingsService{}.ServerConfigChanged(previous, next)
}
