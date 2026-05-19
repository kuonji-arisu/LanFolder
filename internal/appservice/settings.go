package appservice

import (
	"lanfolder/internal/config"
	"lanfolder/internal/desktop"
	"lanfolder/internal/i18n"
	"lanfolder/internal/platform"
	"lanfolder/internal/share"
)

type SettingsService struct{}

func (SettingsService) Normalize(cfg config.Config) config.Config {
	if !cfg.Permission.Valid() {
		cfg.Permission = share.PermissionReadOnly
	}
	cfg.AccessSessionLifetime = share.NormalizeAccessSessionLifetime(cfg.AccessSessionLifetime)
	cfg.Language = i18n.NormalizeLanguage(cfg.Language)
	return cfg
}

func (SettingsService) Validate(cfg config.Config) error {
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return newCommandError(errInvalidPort, nil)
	}
	if cfg.AutoShare && !cfg.AccessApproval {
		return newCommandError(errAccessApprovalRequired, nil)
	}
	return nil
}

func (SettingsService) ApplyPlatformSideEffects(previous, next config.Config) error {
	if next.StartAtLogin == previous.StartAtLogin {
		return nil
	}
	return platform.SetStartAtLogin(next.StartAtLogin)
}

func (SettingsService) ServerConfigChanged(previous config.Config, next config.Config) bool {
	return previous.SharedDir != next.SharedDir ||
		previous.Port != next.Port ||
		previous.Permission != next.Permission ||
		previous.ShowHiddenFiles != next.ShowHiddenFiles ||
		previous.AccessApproval != next.AccessApproval ||
		previous.AccessSessionLifetime != next.AccessSessionLifetime ||
		previous.Language != next.Language
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

	cfg = s.settings.Normalize(cfg)
	if err := s.settings.Validate(cfg); err != nil {
		return s.snapshot(s.config), err
	}
	if err := s.settings.ApplyPlatformSideEffects(s.config, cfg); err != nil {
		return s.snapshot(s.config), err
	}
	wasRunning := s.server.Status().Running
	shouldRestart := wasRunning && s.settings.ServerConfigChanged(s.config, cfg)
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
