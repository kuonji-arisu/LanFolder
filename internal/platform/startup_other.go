//go:build !windows

package platform

import "runtime"

func StartAtLoginEnabled() (bool, error) {
	return false, nil
}

func SetStartAtLogin(enabled bool) error {
	if enabled {
		return ErrUnsupported("开机启动暂不支持 " + runtime.GOOS)
	}
	return nil
}
