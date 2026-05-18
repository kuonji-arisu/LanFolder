package notice

import (
	"context"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

func TestRuntimeReportsUnavailableBeforeStartup(t *testing.T) {
	service := &RuntimeService{}

	err := service.sendNotification(notifications.NotificationOptions{ID: "notice", Title: "LanFolder"})

	if err != ErrNotificationUnavailable {
		t.Fatalf("error = %v, want %v", err, ErrNotificationUnavailable)
	}
}

func TestRuntimeShutdownDisablesNotifications(t *testing.T) {
	service := &RuntimeService{}
	service.setReady(true)

	_ = service.ServiceShutdown()
	err := service.sendNotification(notifications.NotificationOptions{ID: "notice", Title: "LanFolder"})

	if err != ErrNotificationUnavailable {
		t.Fatalf("error = %v, want %v", err, ErrNotificationUnavailable)
	}
}

func TestRuntimeStartupHandlesMissingNotifier(t *testing.T) {
	service := &RuntimeService{}

	if err := service.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
		t.Fatalf("startup error = %v, want nil", err)
	}
	if err := service.sendNotification(notifications.NotificationOptions{ID: "notice", Title: "LanFolder"}); err != ErrNotificationUnavailable {
		t.Fatalf("send error = %v, want %v", err, ErrNotificationUnavailable)
	}
}
