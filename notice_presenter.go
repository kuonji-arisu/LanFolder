package main

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"lanfolder/internal/desktop"
	"lanfolder/internal/i18n"
)

const (
	noticePresentationToast  = "toast"
	noticePresentationSystem = "system"
)

type noticeWindow interface {
	IsFocused() bool
	IsVisible() bool
	Flash(enabled bool)
}

type noticeNotifier interface {
	sendNotification(options notifications.NotificationOptions) error
}

func (s *AppService) PresentNotice(notice desktop.Notice, message string) string {
	s.mu.Lock()
	window := s.window
	notifier := s.notifier
	language := s.config.Language
	s.mu.Unlock()
	return presentNotice(window, notifier, notice, message, language)
}

func presentNotice(window noticeWindow, notifier noticeNotifier, notice desktop.Notice, message, language string) string {
	if window == nil {
		return noticePresentationToast
	}

	visible := window.IsVisible()
	if visible {
		if !window.IsFocused() {
			window.Flash(true)
		}
		return noticePresentationToast
	}

	if notifier == nil {
		return noticePresentationToast
	}

	if err := notifier.sendNotification(notifications.NotificationOptions{
		ID:    fmt.Sprintf("notice-%s", notice.ID),
		Title: noticeNotificationTitle(notice, language),
		Body:  message,
		Data: map[string]interface{}{
			"noticeId": notice.ID,
			"level":    string(notice.Level),
			"source":   string(notice.Source),
		},
	}); err != nil {
		return noticePresentationToast
	}
	return noticePresentationSystem
}

func noticeNotificationTitle(notice desktop.Notice, language string) string {
	switch notice.Level {
	case desktop.NoticeError:
		return i18n.T(language, "notice.errorTitle", nil)
	case desktop.NoticeWarning:
		return i18n.T(language, "notice.warningTitle", nil)
	default:
		return "LanFolder"
	}
}
