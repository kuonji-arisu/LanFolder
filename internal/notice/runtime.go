package notice

import (
	"context"
	"errors"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

var ErrNotificationUnavailable = errors.New("notification_unavailable")

type RuntimeService struct {
	notifier *notifications.NotificationService
	mu       sync.RWMutex
	ready    bool
}

func NewRuntimeService(notifier *notifications.NotificationService) *RuntimeService {
	return &RuntimeService{notifier: notifier}
}

func (s *RuntimeService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
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

func (s *RuntimeService) ServiceShutdown() error {
	s.setReady(false)
	if s.notifier == nil {
		return nil
	}
	return s.notifier.ServiceShutdown()
}

func (s *RuntimeService) sendNotification(options notifications.NotificationOptions) error {
	s.mu.RLock()
	ready := s.ready
	s.mu.RUnlock()
	if !ready {
		return ErrNotificationUnavailable
	}
	if s.notifier == nil {
		return ErrNotificationUnavailable
	}
	return s.notifier.SendNotification(options)
}

func (s *RuntimeService) setReady(ready bool) {
	s.mu.Lock()
	s.ready = ready
	s.mu.Unlock()
}
