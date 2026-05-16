package main

import (
	"context"
	"errors"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

var errNotificationUnavailable = errors.New("notification_unavailable")

type notificationRuntimeService struct {
	notifier *notifications.NotificationService
	mu       sync.RWMutex
	ready    bool
}

func (s *notificationRuntimeService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	if s.notifier == nil {
		s.setReady(false)
		return nil
	}
	if err := s.notifier.ServiceStartup(ctx, options); err != nil {
		s.setReady(false)
		return nil
	}
	s.setReady(true)
	return nil
}

func (s *notificationRuntimeService) ServiceShutdown() error {
	s.setReady(false)
	if s.notifier == nil {
		return nil
	}
	return s.notifier.ServiceShutdown()
}

func (s *notificationRuntimeService) sendNotification(options notifications.NotificationOptions) error {
	s.mu.RLock()
	ready := s.ready
	s.mu.RUnlock()
	if !ready {
		return errNotificationUnavailable
	}
	if s.notifier == nil {
		return errNotificationUnavailable
	}
	return s.notifier.SendNotification(options)
}

func (s *notificationRuntimeService) setReady(ready bool) {
	s.mu.Lock()
	s.ready = ready
	s.mu.Unlock()
}
