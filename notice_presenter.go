package main

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"lanfolder/internal/desktop"
)

const (
	noticePresentationToast     = "toast"
	noticePresentationAttention = "attention"
	noticePresentationSystem    = "system"
)

type noticeWindow interface {
	IsFocused() bool
	IsMinimised() bool
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
	s.mu.Unlock()
	return presentNotice(window, notifier, notice, message)
}

func presentNotice(window noticeWindow, notifier noticeNotifier, notice desktop.Notice, message string) string {
	if window == nil {
		return noticePresentationToast
	}

	visible := window.IsVisible()
	minimised := window.IsMinimised()
	if visible && !minimised && window.IsFocused() {
		return noticePresentationToast
	}

	if visible || minimised {
		window.Flash(true)
		return noticePresentationAttention
	}

	if notifier == nil {
		return noticePresentationToast
	}

	if err := notifier.sendNotification(notifications.NotificationOptions{
		ID:    fmt.Sprintf("notice-%s", notice.ID),
		Title: noticeNotificationTitle(notice),
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

func noticeNotificationTitle(notice desktop.Notice) string {
	switch notice.Level {
	case desktop.NoticeError:
		return "LanFolder 操作失败"
	case desktop.NoticeWarning:
		return "LanFolder 需要注意"
	default:
		return "LanFolder"
	}
}
