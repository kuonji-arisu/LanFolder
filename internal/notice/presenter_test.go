package notice

import (
	"errors"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"lanfolder/internal/desktop"
	"lanfolder/internal/i18n"
)

func TestPresentReturnsToastForFocusedVisibleWindow(t *testing.T) {
	window := &fakeWindow{focused: true, visible: true}
	notifier := &fakeNotifier{}

	got := present(window, notifier, noticeForPresentation(), "hello", i18n.Chinese)

	if got != PresentationToast {
		t.Fatalf("presentation = %q, want %q", got, PresentationToast)
	}
	if len(window.flashes) != 0 {
		t.Fatalf("flashes = %#v, want none", window.flashes)
	}
	if len(notifier.sent) != 0 {
		t.Fatalf("notifications = %#v, want none", notifier.sent)
	}
}

func TestPresentFlashesAndReturnsToastForVisibleUnfocusedWindow(t *testing.T) {
	window := &fakeWindow{visible: true}
	notifier := &fakeNotifier{}

	got := present(window, notifier, noticeForPresentation(), "hello", i18n.Chinese)

	if got != PresentationToast {
		t.Fatalf("presentation = %q, want %q", got, PresentationToast)
	}
	if len(window.flashes) != 1 || !window.flashes[0] {
		t.Fatalf("flashes = %#v, want [true]", window.flashes)
	}
	if len(notifier.sent) != 0 {
		t.Fatalf("notifications = %#v, want none", notifier.sent)
	}
}

func TestPresentSendsSystemNotificationForHiddenWindow(t *testing.T) {
	window := &fakeWindow{}
	notifier := &fakeNotifier{}
	notice := noticeForPresentation()

	got := present(window, notifier, notice, "hello", i18n.Chinese)

	if got != PresentationSystem {
		t.Fatalf("presentation = %q, want %q", got, PresentationSystem)
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

func TestPresentFallsBackToToastWhenSystemNotificationFails(t *testing.T) {
	window := &fakeWindow{}
	notifier := &fakeNotifier{err: errors.New("notification failed")}

	got := present(window, notifier, noticeForPresentation(), "hello", i18n.Chinese)

	if got != PresentationToast {
		t.Fatalf("presentation = %q, want %q", got, PresentationToast)
	}
}

func noticeForPresentation() desktop.Notice {
	return desktop.Notice{
		ID:     "42",
		Level:  desktop.NoticeWarning,
		Source: desktop.NoticeSourceSystem,
	}
}

type fakeWindow struct {
	focused bool
	visible bool
	flashes []bool
}

func (w *fakeWindow) IsFocused() bool {
	return w.focused
}

func (w *fakeWindow) IsVisible() bool {
	return w.visible
}

func (w *fakeWindow) Flash(enabled bool) {
	w.flashes = append(w.flashes, enabled)
}

type fakeNotifier struct {
	sent []notifications.NotificationOptions
	err  error
}

func (n *fakeNotifier) sendNotification(options notifications.NotificationOptions) error {
	n.sent = append(n.sent, options)
	return n.err
}
