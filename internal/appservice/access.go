package appservice

import (
	"sync"
	"time"

	"lanfolder/internal/desktop"
	"lanfolder/internal/i18n"
	"lanfolder/internal/server"
	"lanfolder/internal/share"
)

const accessNoticeCooldown = 10 * time.Second

type AccessService struct {
	mu                 sync.Mutex
	server             *server.Server
	lastAccessNoticeAt time.Time
}

func NewAccessService(server *server.Server) *AccessService {
	return &AccessService{server: server}
}

func (a *AccessService) Pending() []share.AccessRequest {
	return a.server.PendingAccessRequests()
}

func (a *AccessService) Sessions() []share.AccessSession {
	return a.server.AccessSessions()
}

func (a *AccessService) Approve(id string) error {
	return a.server.ApproveAccessRequest(id)
}

func (a *AccessService) Deny(id string) error {
	return a.server.DenyAccessRequest(id)
}

func (a *AccessService) Revoke(id string) bool {
	return a.server.RevokeAccessSession(id)
}

func (a *AccessService) ShouldShowNotice(now time.Time) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.lastAccessNoticeAt.IsZero() && now.Sub(a.lastAccessNoticeAt) < accessNoticeCooldown {
		return false
	}
	a.lastAccessNoticeAt = now
	return true
}

func (s *AppService) PendingAccessRequests() []share.AccessRequest {
	return s.accessService().Pending()
}

func (s *AppService) AccessSessions() []share.AccessSession {
	return s.accessService().Sessions()
}

func (s *AppService) ApproveAccessRequest(id string) error {
	if err := s.accessService().Approve(id); err != nil {
		return newCommandError(errAccessRequestUnavailable, nil)
	}
	s.emitStateChanged("access")
	return nil
}

func (s *AppService) DenyAccessRequest(id string) error {
	if err := s.accessService().Deny(id); err != nil {
		return newCommandError(errAccessRequestUnavailable, nil)
	}
	s.emitStateChanged("access")
	return nil
}

func (s *AppService) RevokeAccessSession(id string) error {
	if ok := s.accessService().Revoke(id); !ok {
		return newCommandError(errAccessRequestUnavailable, nil)
	}
	s.emitStateChanged("access")
	return nil
}

func (s *AppService) handleAccessRequestNotice() {
	s.emitStateChanged("access")
	if !s.accessService().ShouldShowNotice(time.Now()) {
		return
	}
	s.mu.Lock()
	language := s.config.Language
	s.mu.Unlock()
	s.addNotice(desktop.NoticeInfo, desktop.NoticeSourceSystem, nil, i18n.T(language, "notice.accessRequest", nil))
}

func (s *AppService) shouldShowAccessNotice(now time.Time) bool {
	return s.accessService().ShouldShowNotice(now)
}
