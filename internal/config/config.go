package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"lanfolder/internal/share"
)

type Config struct {
	SharedDir       string           `json:"sharedDir"`
	Port            int              `json:"port"`
	Permission      share.Permission `json:"permission"`
	AccessApproval  bool             `json:"accessApproval"`
	AutoShare       bool             `json:"autoShare"`
	StartAtLogin    bool             `json:"startAtLogin"`
	KeepInTray      bool             `json:"keepInTray"`
	ShowHiddenFiles bool             `json:"showHiddenFiles"`
}

func Default() Config {
	return Config{
		Port:       8899,
		Permission: share.PermissionReadOnly,
	}
}

func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "LanFolder", "config.json"), nil
}

func Load() Config {
	cfg := Default()
	p, err := Path()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return cfg
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default()
	}
	if cfg.Port == 0 {
		cfg.Port = 8899
	}
	if !cfg.Permission.Valid() {
		cfg.Permission = share.PermissionReadOnly
	}
	if cfg.AutoShare && !cfg.AccessApproval {
		cfg.AutoShare = false
	}
	return cfg
}

func Save(cfg Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	return saveToPath(p, cfg)
}

func saveToPath(p string, cfg Config) error {
	if cfg.Port == 0 {
		cfg.Port = 8899
	}
	if !cfg.Permission.Valid() {
		cfg.Permission = share.PermissionReadOnly
	}
	if cfg.AutoShare && !cfg.AccessApproval {
		cfg.AutoShare = false
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(p, data, 0600)
}
