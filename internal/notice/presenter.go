package notice

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"lanfolder/internal/desktop"
	"lanfolder/internal/i18n"
)

const (
	PresentationToast  = "toast"
	PresentationSystem = "system"
)

type Window interface {
	IsFocused() bool
	IsVisible() bool
	Flash(enabled bool)
}

type notifier interface {
	sendNotification(options notifications.NotificationOptions) error
}

func Present(window Window, notifier *RuntimeService, notice desktop.Notice, message, language string) string {
	return present(window, notifier, notice, message, language)
}

func present(window Window, notifier notifier, notice desktop.Notice, message, language string) string {
	if window == nil {
		return PresentationToast
	}

	visible := window.IsVisible()
	if visible {
		if !window.IsFocused() {
			window.Flash(true)
		}
		return PresentationToast
	}

	if notifier == nil {
		return PresentationToast
	}

	if err := notifier.sendNotification(notifications.NotificationOptions{
		ID:    fmt.Sprintf("notice-%s", notice.ID),
		Title: notificationTitle(notice, language),
		Body:  message,
		Data: map[string]interface{}{
			"noticeId": notice.ID,
			"level":    string(notice.Level),
			"source":   string(notice.Source),
		},
	}); err != nil {
		return PresentationToast
	}
	return PresentationSystem
}

func notificationTitle(notice desktop.Notice, language string) string {
	switch notice.Level {
	case desktop.NoticeError:
		return i18n.T(language, "notice.errorTitle", nil)
	case desktop.NoticeWarning:
		return i18n.T(language, "notice.warningTitle", nil)
	default:
		return "LanFolder"
	}
}
