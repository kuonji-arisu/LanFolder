package main

import (
	"context"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

func TestNotificationRuntimeReportsUnavailableBeforeStartup(t *testing.T) {
	service := &notificationRuntimeService{}

	err := service.sendNotification(notifications.NotificationOptions{ID: "notice", Title: "LanFolder"})

	if err != errNotificationUnavailable {
		t.Fatalf("error = %v, want %v", err, errNotificationUnavailable)
	}
}

func TestNotificationRuntimeShutdownDisablesNotifications(t *testing.T) {
	service := &notificationRuntimeService{}
	service.setReady(true)

	_ = service.ServiceShutdown()
	err := service.sendNotification(notifications.NotificationOptions{ID: "notice", Title: "LanFolder"})

	if err != errNotificationUnavailable {
		t.Fatalf("error = %v, want %v", err, errNotificationUnavailable)
	}
}

func TestNotificationRuntimeStartupHandlesMissingNotifier(t *testing.T) {
	service := &notificationRuntimeService{}

	if err := service.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
		t.Fatalf("startup error = %v, want nil", err)
	}
	if err := service.sendNotification(notifications.NotificationOptions{ID: "notice", Title: "LanFolder"}); err != errNotificationUnavailable {
		t.Fatalf("send error = %v, want %v", err, errNotificationUnavailable)
	}
}
