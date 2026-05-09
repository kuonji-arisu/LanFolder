//go:build windows

package platform

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const startupAppName = "LanFolder"
const startupRunKey = `Software\Microsoft\Windows\CurrentVersion\Run`

func StartAtLoginEnabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, startupRunKey, registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer key.Close()

	value, _, err := key.GetStringValue(startupAppName)
	if err == registry.ErrNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	exe, err := os.Executable()
	if err != nil {
		return false, err
	}
	return value == quoteCommand(filepath.Clean(exe)), nil
}

func SetStartAtLogin(enabled bool) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, startupRunKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	if !enabled {
		if err := key.DeleteValue(startupAppName); err != nil && err != registry.ErrNotExist {
			return err
		}
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return key.SetStringValue(startupAppName, quoteCommand(filepath.Clean(exe)))
}

func quoteCommand(path string) string {
	return `"` + path + `"`
}
