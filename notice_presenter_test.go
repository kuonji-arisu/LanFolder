package main

import (
	"errors"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"lanfolder/internal/desktop"
)

func TestPresentNoticeReturnsToastForFocusedVisibleWindow(t *testing.T) {
	window := &fakeNoticeWindow{focused: true, visible: true}
	notifier := &fakeNoticeNotifier{}

	got := presentNotice(window, notifier, noticeForPresentation(), "hello")

	if got != noticePresentationToast {
		t.Fatalf("presentation = %q, want %q", got, noticePresentationToast)
	}
	if len(window.flashes) != 0 {
		t.Fatalf("flashes = %#v, want none", window.flashes)
	}
	if len(notifier.sent) != 0 {
		t.Fatalf("notifications = %#v, want none", notifier.sent)
	}
}

func TestPresentNoticeFlashesAndReturnsToastForVisibleUnfocusedWindow(t *testing.T) {
	window := &fakeNoticeWindow{visible: true}
	notifier := &fakeNoticeNotifier{}

	got := presentNotice(window, notifier, noticeForPresentation(), "hello")

	if got != noticePresentationToast {
		t.Fatalf("presentation = %q, want %q", got, noticePresentationToast)
	}
	if len(window.flashes) != 1 || !window.flashes[0] {
		t.Fatalf("flashes = %#v, want [true]", window.flashes)
	}
	if len(notifier.sent) != 0 {
		t.Fatalf("notifications = %#v, want none", notifier.sent)
	}
}

func TestPresentNoticeSendsSystemNotificationForHiddenWindow(t *testing.T) {
	window := &fakeNoticeWindow{}
	notifier := &fakeNoticeNotifier{}
	notice := noticeForPresentation()

	got := presentNotice(window, notifier, notice, "hello")

	if got != noticePresentationSystem {
		t.Fatalf("presentation = %q, want %q", got, noticePresentationSystem)
	}
	if len(window.flashes) != 0 {
		t.Fatalf("flashes = %#v, want none", window.flashes)
	}
	if len(notifier.sent) != 1 {
		t.Fatalf("notifications = %d, want 1", len(notifier.sent))
	}
	sent := notifier.sent[0]
	if sent.ID != "notice-42" || sent.Title != "LanFolder 需要注意" || sent.Body != "hello" {
		t.Fatalf("notification = %#v", sent)
	}
	if sent.Data["noticeId"] != "42" || sent.Data["level"] != string(desktop.NoticeWarning) || sent.Data["source"] != string(desktop.NoticeSourceSystem) {
		t.Fatalf("notification data = %#v", sent.Data)
	}
}

func TestPresentNoticeFallsBackToToastWhenSystemNotificationFails(t *testing.T) {
	window := &fakeNoticeWindow{}
	notifier := &fakeNoticeNotifier{err: errors.New("notification failed")}

	got := presentNotice(window, notifier, noticeForPresentation(), "hello")

	if got != noticePresentationToast {
		t.Fatalf("presentation = %q, want %q", got, noticePresentationToast)
	}
}

func noticeForPresentation() desktop.Notice {
	return desktop.Notice{
		ID:     "42",
		Level:  desktop.NoticeWarning,
		Source: desktop.NoticeSourceSystem,
	}
}

type fakeNoticeWindow struct {
	focused bool
	visible bool
	flashes []bool
}

func (w *fakeNoticeWindow) IsFocused() bool {
	return w.focused
}

func (w *fakeNoticeWindow) IsVisible() bool {
	return w.visible
}

func (w *fakeNoticeWindow) Flash(enabled bool) {
	w.flashes = append(w.flashes, enabled)
}

type fakeNoticeNotifier struct {
	sent []notifications.NotificationOptions
	err  error
}

func (n *fakeNoticeNotifier) sendNotification(options notifications.NotificationOptions) error {
	n.sent = append(n.sent, options)
	return n.err
}
